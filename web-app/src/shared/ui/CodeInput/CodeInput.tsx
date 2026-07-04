import { useEffect, useRef, useCallback, type KeyboardEvent, type ClipboardEvent } from "react";

import styles from "./CodeInput.module.css";

interface CodeInputProps {
  length?: number;
  value: string;
  onChange: (value: string) => void;
  mode?: "numeric" | "alphanumeric";
  separatorAfter?: number;
  autoFocus?: boolean;
}

export function CodeInput({
  length = 6,
  value,
  onChange,
  mode = "numeric",
  separatorAfter,
  autoFocus = false,
}: CodeInputProps) {
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  useEffect(() => {
    if (autoFocus) inputRefs.current[0]?.focus();
  }, [autoFocus]);

  const sanitize = useCallback(
    (val: string) => {
      const cleaned =
        mode === "alphanumeric"
          ? val.replace(/[^a-zA-Z0-9]/g, "").toUpperCase()
          : val.replace(/\D/g, "");
      return cleaned.slice(0, length).split("");
    },
    [length, mode],
  );

  const getCells = useCallback(
    (val: string) => {
      const cells = sanitize(val);
      while (cells.length < length) cells.push("");
      return cells;
    },
    [length, sanitize],
  );

  const cells = getCells(value);

  const handleChange = useCallback(
    (index: number, val: string) => {
      const ch = sanitize(val).join("").slice(-1);
      const next = [...cells];
      next[index] = ch;
      const joined = next.join("");
      onChange(joined);

      if (ch && index < length - 1) {
        inputRefs.current[index + 1]?.focus();
      }
    },
    [cells, length, onChange, sanitize],
  );

  const handleKeyDown = useCallback(
    (index: number, e: KeyboardEvent) => {
      if (e.key === "Backspace" && !cells[index] && index > 0) {
        const next = [...cells];
        next[index - 1] = "";
        onChange(next.join(""));
        inputRefs.current[index - 1]?.focus();
      }
    },
    [cells, onChange],
  );

  const handlePaste = useCallback(
    (e: ClipboardEvent) => {
      e.preventDefault();
      const text = sanitize(e.clipboardData.getData("text")).join("");
      const next = text.split("");
      while (next.length < length) next.push("");
      onChange(next.join(""));
      const focusIdx = Math.min(text.length, length - 1);
      inputRefs.current[focusIdx]?.focus();
    },
    [length, onChange, sanitize],
  );

  const renderCell = (i: number) => (
    <input
      key={i}
      ref={(el) => {
        inputRefs.current[i] = el;
      }}
      type="text"
      inputMode={mode === "alphanumeric" ? "text" : "numeric"}
      autoCapitalize={mode === "alphanumeric" ? "characters" : undefined}
      maxLength={1}
      aria-label={`Символ ${i + 1}`}
      className={styles.cell}
      value={cells[i]}
      onChange={(e) => handleChange(i, e.target.value)}
      onKeyDown={(e) => handleKeyDown(i, e)}
      onPaste={handlePaste}
    />
  );

  // No separator: render all cells in one group
  if (!separatorAfter) {
    return (
      <div className={styles.wrapper}>
        <div className={styles.group}>{cells.map((_, i) => renderCell(i))}</div>
      </div>
    );
  }

  // With separator: split into two groups
  const firstGroup = Array.from({ length: separatorAfter }, (_, i) => i);
  const secondGroup = Array.from({ length: length - separatorAfter }, (_, i) => i + separatorAfter);

  return (
    <div className={styles.wrapper}>
      <div className={styles.group}>{firstGroup.map((i) => renderCell(i))}</div>

      <span className={styles.separator} aria-hidden="true">
        -
      </span>

      <div className={styles.group}>{secondGroup.map((i) => renderCell(i))}</div>
    </div>
  );
}
