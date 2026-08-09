export const MOBILELAB_PROTOCOL_VERSION = 1 as const;

export type MobileLabAttributes = Record<string, unknown>;
export type MobileLabEventKind = 'lifecycle' | 'marker' | 'assertion';

export interface MobileLabEvent {
  protocolVersion: typeof MOBILELAB_PROTOCOL_VERSION;
  framework: 'capacitor';
  kind: MobileLabEventKind;
  name: string;
  passed?: boolean;
  sessionId?: string;
  attributes?: MobileLabAttributes;
}

export interface MobileLabTransport {
  send(event: MobileLabEvent): Promise<void>;
}

export interface MobileLabClientOptions {
  endpoint: string | URL;
  sessionId?: string;
  timeoutMs?: number;
  transport?: MobileLabTransport;
  fetchImplementation?: typeof fetch;
}

export class MobileLabHttpTransport implements MobileLabTransport {
  readonly eventsUrl: URL;
  readonly timeoutMs: number;
  readonly fetchImplementation: typeof fetch;

  constructor(
    endpoint: string | URL,
    options: { timeoutMs?: number; fetchImplementation?: typeof fetch } = {},
  ) {
    this.eventsUrl = new URL('/__mobilelab/sdk/events', endpoint);
    this.timeoutMs = options.timeoutMs ?? 3000;
    this.fetchImplementation = options.fetchImplementation ?? globalThis.fetch;
    if (typeof this.fetchImplementation !== 'function') {
      throw new MobileLabError('A fetch implementation is required');
    }
  }

  async send(event: MobileLabEvent): Promise<void> {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), this.timeoutMs);
    try {
      const response = await this.fetchImplementation(this.eventsUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-MobileLab-SDK': '@mobilelab/capacitor/0.5.0',
        },
        body: JSON.stringify(event),
        signal: controller.signal,
      });
      if (!response.ok) {
        throw new MobileLabError(`SDK bridge returned HTTP ${response.status}`);
      }
    } catch (error) {
      if (error instanceof MobileLabError) throw error;
      throw new MobileLabError(`SDK bridge is unreachable at ${this.eventsUrl}`, error);
    } finally {
      clearTimeout(timeout);
    }
  }
}

export class MobileLabClient {
  readonly sessionId: string | undefined;
  readonly transport: MobileLabTransport;

  constructor(options: MobileLabClientOptions) {
    this.sessionId = options.sessionId;
    this.transport = options.transport ?? new MobileLabHttpTransport(options.endpoint, {
      ...(options.timeoutMs === undefined ? {} : { timeoutMs: options.timeoutMs }),
      ...(options.fetchImplementation === undefined
        ? {}
        : { fetchImplementation: options.fetchImplementation }),
    });
  }

  lifecycle(name: string, attributes: MobileLabAttributes = {}): Promise<void> {
    return this.send('lifecycle', name, undefined, attributes);
  }

  marker(name: string, attributes: MobileLabAttributes = {}): Promise<void> {
    return this.send('marker', name, undefined, attributes);
  }

  assertThat(name: string, passed: boolean, attributes: MobileLabAttributes = {}): Promise<void> {
    return this.send('assertion', name, passed, attributes);
  }

  private send(
    kind: MobileLabEventKind,
    name: string,
    passed: boolean | undefined,
    attributes: MobileLabAttributes,
  ): Promise<void> {
    return this.transport.send({
      protocolVersion: MOBILELAB_PROTOCOL_VERSION,
      framework: 'capacitor',
      kind,
      name,
      ...(passed === undefined ? {} : { passed }),
      ...(this.sessionId ? { sessionId: this.sessionId } : {}),
      ...(Object.keys(attributes).length === 0 ? {} : { attributes: { ...attributes } }),
    });
  }
}

export interface CapacitorAppState {
  isActive: boolean;
}

export interface CapacitorPluginListenerHandle {
  remove(): Promise<void>;
}

export interface CapacitorAppLike {
  addListener(
    eventName: 'appStateChange',
    listener: (state: CapacitorAppState) => void,
  ): Promise<CapacitorPluginListenerHandle>;
}

export class MobileLabLifecycleReporter {
  private listener: CapacitorPluginListenerHandle | undefined;
  private attaching: Promise<void> | undefined;
  private lastLifecycle: string | undefined;

  constructor(
    readonly client: MobileLabClient,
    readonly app: CapacitorAppLike,
    readonly onError?: (error: unknown) => void,
  ) {}

  attach(reportReady = true): Promise<void> {
    if (this.listener) return Promise.resolve();
    if (this.attaching) return this.attaching;
    const operation = this.app.addListener('appStateChange', (state) => this.onState(state))
      .then((listener) => {
        this.listener = listener;
        if (reportReady) this.report(this.client.lifecycle('ready'));
      })
      .finally(() => {
        this.attaching = undefined;
      });
    this.attaching = operation;
    return operation;
  }

  async detach(): Promise<void> {
    await this.attaching;
    const listener = this.listener;
    this.listener = undefined;
    if (listener) await listener.remove();
  }

  private onState(state: CapacitorAppState): void {
    const lifecycle = state.isActive ? 'foreground' : 'background';
    if (lifecycle === this.lastLifecycle) return;
    this.lastLifecycle = lifecycle;
    this.report(this.client.lifecycle(lifecycle, { capacitorIsActive: state.isActive }));
  }

  private report(operation: Promise<void>): void {
    void operation.catch((error: unknown) => this.onError?.(error));
  }
}

export class MobileLabError extends Error {
  constructor(message: string, readonly cause?: unknown) {
    super(message);
    this.name = 'MobileLabError';
  }
}
