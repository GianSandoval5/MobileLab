package datastore

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const maxDocumentBytes = 1 << 20

type Handler struct {
	config Config
	store  *Store
}

func NewHandler(cfg Config, store *Store) (*Handler, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if store == nil {
		return nil, errors.New("business data store is required")
	}
	return &Handler{config: cfg, store: store}, nil
}

// Match implements sandbox.DynamicHandler. It matches a resource path even
// when the method is unsupported so the data API can return HTTP 405.
func (h *Handler) Match(_ string, requestPath string) (bool, bool) {
	for _, name := range h.config.Names() {
		resource := h.config.Resources[name]
		if requestPath == resource.Path || (!resource.Singleton && itemID(resource.Path, requestPath) != "") {
			return resource.Protected, true
		}
	}
	return false, false
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	name, resource, id, found := h.find(request.URL.Path)
	if !found {
		writeDataJSON(writer, http.StatusNotFound, map[string]any{"error": "database resource not found"})
		return
	}
	if resource.Singleton {
		h.serveSingleton(writer, request, name, resource)
		return
	}
	h.serveCollection(writer, request, name, resource, id)
}

func (h *Handler) find(requestPath string) (string, ResourceDefinition, string, bool) {
	for _, name := range h.config.Names() {
		resource := h.config.Resources[name]
		if requestPath == resource.Path {
			return name, resource, "", true
		}
		if !resource.Singleton {
			if id := itemID(resource.Path, requestPath); id != "" {
				return name, resource, id, true
			}
		}
	}
	return "", ResourceDefinition{}, "", false
}

func itemID(base, requestPath string) string {
	if !strings.HasPrefix(requestPath, base+"/") {
		return ""
	}
	raw := strings.TrimPrefix(requestPath, base+"/")
	if raw == "" || strings.Contains(raw, "/") {
		return ""
	}
	id, err := url.PathUnescape(raw)
	if err != nil || id == "" || strings.Contains(id, "/") {
		return ""
	}
	return id
}

func (h *Handler) serveSingleton(writer http.ResponseWriter, request *http.Request, name string, resource ResourceDefinition) {
	switch request.Method {
	case http.MethodGet:
		document, found, err := h.store.Get(request.Context(), name, singletonID)
		if err != nil {
			h.internalError(writer, err)
			return
		}
		if !found {
			writeDataJSON(writer, http.StatusNotFound, map[string]any{"error": "resource is empty", "resource": name})
			return
		}
		writeDataJSON(writer, http.StatusOK, document)
	case http.MethodPut, http.MethodPatch:
		document, err := decodeDocument(request)
		if err != nil {
			writeDataJSON(writer, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		if request.Method == http.MethodPatch {
			current, found, getErr := h.store.Get(request.Context(), name, singletonID)
			if getErr != nil {
				h.internalError(writer, getErr)
				return
			}
			if found {
				document = merge(current, document)
			}
		}
		if err := h.store.Put(request.Context(), name, singletonID, document); err != nil {
			h.internalError(writer, err)
			return
		}
		writeDataJSON(writer, http.StatusOK, document)
	default:
		writer.Header().Set("Allow", "GET, PUT, PATCH")
		writeDataJSON(writer, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}
}

func (h *Handler) serveCollection(writer http.ResponseWriter, request *http.Request, name string, resource ResourceDefinition, id string) {
	if id == "" {
		switch request.Method {
		case http.MethodGet:
			documents, err := h.store.List(request.Context(), name)
			if err != nil {
				h.internalError(writer, err)
				return
			}
			writeDataJSON(writer, http.StatusOK, documents)
		case http.MethodPost:
			document, err := decodeDocument(request)
			if err != nil {
				writeDataJSON(writer, http.StatusBadRequest, map[string]any{"error": err.Error()})
				return
			}
			id, _ := document[resource.ID].(string)
			if value, exists := document[resource.ID]; exists && value != "" && id == "" {
				writeDataJSON(writer, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("field %q must be a string", resource.ID)})
				return
			}
			if id == "" {
				id, err = randomID()
				if err != nil {
					h.internalError(writer, err)
					return
				}
				document[resource.ID] = id
			}
			if _, exists, getErr := h.store.Get(request.Context(), name, id); getErr != nil {
				h.internalError(writer, getErr)
				return
			} else if exists {
				writeDataJSON(writer, http.StatusConflict, map[string]any{"error": "resource already exists", "id": id})
				return
			}
			if err := h.store.Create(request.Context(), name, id, document); err != nil {
				h.internalError(writer, err)
				return
			}
			writeDataJSON(writer, http.StatusCreated, document)
		default:
			writer.Header().Set("Allow", "GET, POST")
			writeDataJSON(writer, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		}
		return
	}

	switch request.Method {
	case http.MethodGet:
		document, found, err := h.store.Get(request.Context(), name, id)
		if err != nil {
			h.internalError(writer, err)
			return
		}
		if !found {
			writeDataJSON(writer, http.StatusNotFound, map[string]any{"error": "resource not found", "id": id})
			return
		}
		writeDataJSON(writer, http.StatusOK, document)
	case http.MethodPut, http.MethodPatch:
		document, err := decodeDocument(request)
		if err != nil {
			writeDataJSON(writer, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		if bodyID, exists := document[resource.ID]; exists && bodyID != id {
			writeDataJSON(writer, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("field %q must match path id %q", resource.ID, id)})
			return
		}
		document[resource.ID] = id
		if request.Method == http.MethodPatch {
			current, found, getErr := h.store.Get(request.Context(), name, id)
			if getErr != nil {
				h.internalError(writer, getErr)
				return
			}
			if !found {
				writeDataJSON(writer, http.StatusNotFound, map[string]any{"error": "resource not found", "id": id})
				return
			}
			document = merge(current, document)
		}
		if err := h.store.Put(request.Context(), name, id, document); err != nil {
			h.internalError(writer, err)
			return
		}
		writeDataJSON(writer, http.StatusOK, document)
	case http.MethodDelete:
		deleted, err := h.store.Delete(request.Context(), name, id)
		if err != nil {
			h.internalError(writer, err)
			return
		}
		if !deleted {
			writeDataJSON(writer, http.StatusNotFound, map[string]any{"error": "resource not found", "id": id})
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	default:
		writer.Header().Set("Allow", "GET, PUT, PATCH, DELETE")
		writeDataJSON(writer, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}
}

func decodeDocument(request *http.Request) (map[string]any, error) {
	decoder := json.NewDecoder(io.LimitReader(request.Body, maxDocumentBytes+1))
	decoder.UseNumber()
	var document map[string]any
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("body must be one JSON object: %w", err)
	}
	if document == nil {
		return nil, errors.New("body must be one JSON object")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, errors.New("body must contain exactly one JSON object no larger than 1 MiB")
	}
	return document, nil
}

func merge(base, patch map[string]any) map[string]any {
	merged := make(map[string]any, len(base)+len(patch))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range patch {
		merged[key] = value
	}
	return merged
}

func randomID() (string, error) {
	data := make([]byte, 12)
	if _, err := rand.Read(data); err != nil {
		return "", fmt.Errorf("generate resource id: %w", err)
	}
	return hex.EncodeToString(data), nil
}

func (h *Handler) internalError(writer http.ResponseWriter, err error) {
	writeDataJSON(writer, http.StatusInternalServerError, map[string]any{"error": "business database operation failed"})
}

func writeDataJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if value != nil {
		_ = json.NewEncoder(writer).Encode(value)
	}
}
