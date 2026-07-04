import { useCallback, useRef, useState } from "react";
import { Icon } from "@iconify/react";
import styles from "./ThemeToggle.module.css";

function getInitialTheme(): string {
  if (typeof document === "undefined") return "light";
  return document.documentElement.getAttribute("data-theme") || "light";
}

export function ThemeToggle() {
  const [isDark, setIsDark] = useState(() => getInitialTheme() === "dark");
  const buttonRef = useRef<HTMLButtonElement>(null);

  const toggle = useCallback(
    (e: React.MouseEvent<HTMLButtonElement>) => {
      const next = isDark ? "light" : "dark";

      const applyTheme = () => {
        document.documentElement.setAttribute("data-theme", next);
        localStorage.setItem("cd-theme", next);
        setIsDark(!isDark);
      };
      const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

      if (!reduceMotion && typeof document.startViewTransition === "function") {
        const x = e.clientX;
        const y = e.clientY;
        const endRadius = Math.hypot(
          Math.max(x, window.innerWidth - x),
          Math.max(y, window.innerHeight - y),
        );

        const transition = document.startViewTransition(applyTheme);
        transition.ready
          .then(() => {
            document.documentElement.animate(
              [
                { clipPath: `circle(0px at ${x}px ${y}px)` },
                { clipPath: `circle(${endRadius}px at ${x}px ${y}px)` },
              ],
              {
                duration: 500,
                easing: "cubic-bezier(.4,0,.2,1)",
                pseudoElement: "::view-transition-new(root)",
              },
            );
          })
          .catch(() => {});
      } else {
        applyTheme();
      }
    },
    [isDark],
  );

  return (
    <button
      ref={buttonRef}
      className={styles.toggle}
      onClick={toggle}
      title={isDark ? "Светлая тема" : "Тёмная тема"}
      type="button"
    >
      <span className={`${styles.iconWrap} ${isDark ? styles.dark : styles.light}`}>
        {isDark ? <Icon icon="tabler:moon" width={20} /> : <Icon icon="tabler:sun" width={20} />}
      </span>
    </button>
  );
}
