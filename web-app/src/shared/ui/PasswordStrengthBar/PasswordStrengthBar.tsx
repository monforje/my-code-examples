import styles from "./PasswordStrengthBar.module.css";

interface PasswordStrengthBarProps {
  password: string;
}

const barColors = [
  "var(--cd-gray-300)",
  "var(--cd-red-600)",
  "var(--cd-yellow-600)",
  "#9bc158",
  "var(--cd-green-600)",
];

export function PasswordStrengthBar({ password }: PasswordStrengthBarProps) {
  const score = getPasswordScore(password);
  const width = password.length === 0 ? 0 : (score / 4) * 100;

  return (
    <div className={styles.track}>
      <div
        className={styles.fill}
        style={{
          width: `${width}%`,
          background: barColors[score],
        }}
      />
    </div>
  );
}

function getPasswordScore(password: string): number {
  if (!password) return 0;

  let score = 0;
  if (password.length >= 8) score++;
  if (password.length >= 12) score++;
  if (/[a-z]/.test(password) && /[A-Z]/.test(password)) score++;
  if (/\d/.test(password)) score++;
  if (/[^A-Za-z0-9]/.test(password)) score++;

  return Math.min(score, 4);
}
