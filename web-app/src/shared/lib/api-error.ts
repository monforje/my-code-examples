import { ErrorCode } from "@shared/api";

interface ApiErrorLike {
  status?: number;
  data?: {
    code?: string;
    message?: string;
  };
}

export function getApiErrorMessage(
  error: unknown,
  fallback = "Что-то пошло не так. Попробуйте ещё раз.",
): string {
  const apiError = error as ApiErrorLike;

  if (apiError.data?.code) {
    switch (apiError.data.code) {
      case ErrorCode.VALIDATION_ERROR:
      case ErrorCode.INVALID_JSON:
        return "Проверьте введённые данные.";
      case ErrorCode.MISSING_AUTH_TOKEN:
      case ErrorCode.INVALID_AUTH_TOKEN:
      case ErrorCode.EXPIRED_AUTH_TOKEN:
      case ErrorCode.INVALID_REFRESH_TOKEN:
      case ErrorCode.EXPIRED_REFRESH_TOKEN:
        return "Сессия истекла. Войдите снова.";
      case ErrorCode.INVALID_CREDENTIALS:
        return "Email или пароль указаны неверно.";
      case ErrorCode.EMAIL_NOT_VERIFIED:
        return "Подтвердите email, чтобы продолжить.";
      case ErrorCode.EMAIL_ALREADY_EXISTS:
        return "Аккаунт с таким email уже существует.";
      case ErrorCode.INVALID_CODE:
        return "Неверный код подтверждения.";
      case ErrorCode.EXPIRED_CODE:
        return "Код истёк. Запросите новый.";
      case ErrorCode.TOO_MANY_ATTEMPTS:
        return "Слишком много попыток. Попробуйте позже.";
      case ErrorCode.RATE_LIMITED:
        return "Слишком много запросов. Попробуйте позже.";
      case ErrorCode.CURRENT_PASSWORD_INCORRECT:
        return "Текущий пароль указан неверно.";
      case ErrorCode.RESET_TOKEN_INVALID:
        return "Ссылка для сброса пароля недействительна.";
      case ErrorCode.RESET_TOKEN_EXPIRED:
        return "Ссылка для сброса пароля истекла.";
      case ErrorCode.NOT_FOUND:
        return "Запрос не найден или уже недействителен.";
      case ErrorCode.INTERNAL_ERROR:
        return "Ошибка сервера. Попробуйте позже.";
      case "AVATAR_TOO_LARGE":
        return "Файл слишком большой. Максимальный размер — 5 МБ.";
      case "INVALID_AVATAR_FORMAT":
        return "Поддерживаются только JPEG, PNG и WebP.";
    }
  }

  if (apiError.status) {
    switch (apiError.status) {
      case 400:
      case 422:
        return "Проверьте введённые данные.";
      case 401:
        return "Сессия истекла. Войдите снова.";
      case 403:
        return "Действие недоступно для этого аккаунта.";
      case 404:
        return "Запрос не найден или уже недействителен.";
      case 409:
        return "Данные конфликтуют с уже существующими.";
      case 413:
        return "Файл слишком большой.";
      case 429:
        return "Слишком много запросов. Попробуйте позже.";
      case 500:
        return "Ошибка сервера. Попробуйте позже.";
    }
  }

  return fallback;
}
