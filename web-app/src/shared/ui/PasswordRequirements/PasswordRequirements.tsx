import styles from "./PasswordRequirements.module.css";

interface PasswordRequirementsProps {
  password: string;
}

export function PasswordRequirements({ password }: PasswordRequirementsProps) {
  const checks = [
    { label: "Минимум 8 символов", met: password.length >= 8 },
    { label: "Заглавная буква", met: /[A-ZА-ЯЁ]/.test(password) },
    { label: "Цифра", met: /[0-9]/.test(password) },
  ];

  return (
    <div className={styles.list}>
      {checks.map((c) => (
        <div key={c.label} className={`${styles.item} ${c.met ? styles.met : ""}`}>
          {c.label}
        </div>
      ))}
    </div>
  );
}
