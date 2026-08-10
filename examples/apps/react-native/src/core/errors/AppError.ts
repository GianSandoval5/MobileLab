export class AppError extends Error {
  constructor(message: string, readonly statusCode?: number) {
    super(message);
    this.name = 'AppError';
  }
}

export const errorMessage = (error: unknown): string =>
  error instanceof Error ? error.message : 'Ocurrió un error inesperado.';
