import type { ReactNode } from "react";
import styles from "./SegmentedControl.module.css";

interface SegmentedControlProps {
  value: string;
  options: { value: string; label: ReactNode }[];
  onChange: (value: string) => void;
}

export function SegmentedControl({ value, options, onChange }: SegmentedControlProps) {
  const activeIndex = Math.max(
    0,
    options.findIndex((o) => o.value === value),
  );

  return (
    <div
      className={styles.segmented}
      role="tablist"
      style={{ "--segments": options.length, "--active": activeIndex } as React.CSSProperties}
    >
      {options.map((opt) => {
        const active = opt.value === value;
        return (
          <button
            key={opt.value}
            type="button"
            role="tab"
            aria-selected={active}
            className={`${styles.item} ${active ? styles.active : ""}`}
            onClick={() => {
              if (!active) onChange(opt.value);
            }}
          >
            {opt.label}
          </button>
        );
      })}
    </div>
  );
}
