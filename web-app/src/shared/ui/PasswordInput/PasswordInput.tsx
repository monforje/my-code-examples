import { useState, forwardRef, type InputHTMLAttributes } from "react";
import { Icon } from "@iconify/react";
import styles from "./PasswordInput.module.css";

interface PasswordInputProps extends InputHTMLAttributes<HTMLInputElement> {}

export const PasswordInput = forwardRef<HTMLInputElement, PasswordInputProps>(
  ({ className, ...props }, ref) => {
    const [visible, setVisible] = useState(false);

    return (
      <div className={styles.wrapper}>
        <input
          ref={ref}
          type={visible ? "text" : "password"}
          className={`${styles.input} ${className ?? ""}`}
          {...props}
        />
        <button
          type="button"
          className={styles.toggle}
          onClick={() => setVisible((v) => !v)}
          tabIndex={-1}
          aria-label={visible ? "Скрыть пароль" : "Показать пароль"}
        >
          <Icon icon={visible ? "tabler:eye" : "tabler:eye-off"} width={20} />
        </button>
      </div>
    );
  },
);

PasswordInput.displayName = "PasswordInput";
