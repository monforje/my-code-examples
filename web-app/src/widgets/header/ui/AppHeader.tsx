import { useState, useEffect, useRef } from "react";
import { Link, useNavigate, useLocation } from "react-router";
import { useAuth } from "@features/auth";
import { Avatar, useProfile } from "@entities/user";
import { useToast } from "@shared/lib/use-toast";
import { getInitials } from "@shared/lib/utils";
import { ThemeToggle } from "@features/theme-toggle";
import logo from "@shared/assets/logo.svg";
import styles from "./Header.module.css";

const NAV_ITEMS = [
  { id: "tasks", label: "Задачи", path: "/tasks" },
  { id: "sandbox", label: "SandBox", path: "/sandbox/tasks" },
  { id: "dashboard", label: "Дашборд", path: "#" },
  { id: "courses", label: "Курсы", path: "#" },
  { id: "community", label: "Сообщество", path: "#" },
];

export function AppHeader() {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuth();
  const { data: profile } = useProfile(!!user);
  const { addToast } = useToast();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const email = profile?.email ?? user?.email ?? "";
  const displayName = profile?.display_name || email;
  const avatarUrl = profile?.avatar_url;
  const initials = getInitials(displayName || email);

  const currentPage =
    NAV_ITEMS.find((item) => item.path !== "#" && location.pathname.startsWith(item.path))?.id ||
    "tasks";

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false);
      }
    };
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        setDropdownOpen(false);
      }
    };
    document.addEventListener("click", handleClick);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("click", handleClick);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

  const handleLogout = async () => {
    await logout();
    addToast("Вы вышли из аккаунта", "info");
  };

  return (
    <header className={styles.header}>
      <div className={styles.left}>
        <Link to="/tasks" className={styles.logo}>
          <img src={logo} alt="Codurity" height={32} />
        </Link>
      </div>
      <nav className={styles.nav}>
        {NAV_ITEMS.map((item) => {
          const active = currentPage === item.id;
          return (
            <Link
              key={item.id}
              to={item.path}
              className={`${styles.navItem} ${active ? styles.navItemActive : ""}`}
              onClick={(event) => {
                if (item.path === "#") {
                  event.preventDefault();
                  addToast("Этот раздел пока не реализован", "info");
                }
              }}
            >
              {item.label}
            </Link>
          );
        })}
      </nav>
      <div className={styles.right}>
        <button
          className={styles.searchBtn}
          title="Поиск (Ctrl+K)"
          onClick={() => addToast("Поиск в разработке", "info")}
        >
          <svg
            className={styles.searchIcon}
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="11" cy="11" r="8" />
            <path d="m21 21-4.3-4.3" />
          </svg>
          <span className={styles.searchText}>Поиск...</span>
          <span className={styles.searchShortcut}>⌘K</span>
        </button>

        <div className={styles.notifications}>
          <button
            className={styles.notifBtn}
            title="Уведомления"
            onClick={() => addToast("Уведомления в разработке", "info")}
          >
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9" />
              <path d="M10.3 21a1.94 1.94 0 0 0 3.4 0" />
            </svg>
            <span className={styles.notifBadge} />
          </button>
        </div>

        <div className={styles.user} ref={dropdownRef}>
          <button className={styles.userBtn} onClick={() => setDropdownOpen((o) => !o)}>
            <Avatar className={styles.avatar} src={avatarUrl} initials={initials} />
          </button>
          {dropdownOpen && (
            <div className={styles.dropdown}>
              <div className={styles.dropdownHeader}>
                <Avatar className={styles.dropdownAvatar} src={avatarUrl} initials={initials} />
                <div className={styles.dropdownEmail}>{displayName}</div>
              </div>
              <div className={styles.dropdownItems}>
                <button
                  className={styles.dropdownItem}
                  onClick={() => {
                    navigate("/profile");
                    setDropdownOpen(false);
                  }}
                >
                  Профиль
                </button>
                <button
                  className={styles.dropdownItem}
                  onClick={() => {
                    navigate("/settings/profile");
                    setDropdownOpen(false);
                  }}
                >
                  Настройки
                </button>
                <div className={styles.dropdownDivider} />
                <div className={styles.dropdownThemeToggle}>
                  <span>Тема</span>
                  <ThemeToggle />
                </div>
                <div className={styles.dropdownDivider} />
                <button
                  className={styles.dropdownItemDanger}
                  onClick={() => {
                    handleLogout();
                    setDropdownOpen(false);
                  }}
                >
                  Выйти
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
