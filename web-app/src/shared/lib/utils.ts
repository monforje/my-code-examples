export function formatDate(date: string | Date): string {
  return new Date(date).toLocaleDateString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  });
}

export function getInitials(email: string): string {
  return email[0]?.toUpperCase() ?? "U";
}
