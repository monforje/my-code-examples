import { useState, useRef, useEffect, type ReactNode } from "react";
import styles from "./Dropdown.module.css";

interface DropdownOption {
  value: string;
  label: string;
  hint?: string;
  dotColor?: string;
}

interface DropdownProps {
  value: string;
  options: DropdownOption[];
  placeholder?: string;
  onChange: (value: string) => void;
  renderTrigger?: (label: string, dotColor?: string) => ReactNode;
}

export function Dropdown({
  value,
  options,
  placeholder = "Выбрать",
  onChange,
  renderTrigger: _renderTrigger,
}: DropdownProps) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  const active = options.find((o) => o.value === value);
  const label = active?.label || placeholder;
  const dotColor = active?.dotColor;

  useEffect(() => {
    if (!open) return;
    const handleClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("click", handleClick, { capture: true });
    document.addEventListener("keydown", handleKey);
    return () => {
      document.removeEventListener("click", handleClick, { capture: true });
      document.removeEventListener("keydown", handleKey);
    };
  }, [open]);

  const isActive = value !== "all" && value !== "";

  return (
    <div className={styles.dropdown} ref={ref}>
      <button
        type="button"
        className={`${styles.trigger} ${isActive ? styles.triggerActive : ""}`}
        onClick={() => setOpen((o) => !o)}
      >
        {dotColor && <span className={styles.dot} style={{ background: dotColor }} />}
        <span>{label}</span>
        <svg
          className={styles.chevron}
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </button>
      {open && (
        <div className={styles.menu}>
          {options.map((opt) => {
            const isActive = opt.value === value;
            return (
              <button
                key={opt.value}
                type="button"
                className={`${styles.menuItem} ${isActive ? styles.menuItemActive : ""}`}
                onClick={() => {
                  onChange(opt.value);
                  setOpen(false);
                }}
              >
                {opt.dotColor && (
                  <span className={styles.menuItemDot} style={{ background: opt.dotColor }} />
                )}
                <span className={styles.menuItemLabel}>{opt.label}</span>
                {opt.hint && <span className={styles.menuItemHint}>{opt.hint}</span>}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
