import { Link } from "react-router";
import logo from "@shared/assets/logo.svg";

export function Header() {
  return (
    <header className="cd-container cd-row-between" style={{ padding: "var(--cd-space-5) 0" }}>
      <Link to="/">
        <img src={logo} alt="Codurity" height="32" />
      </Link>
      <nav className="cd-row">
        <Link to="/" className="cd-button cd-button-ghost">
          Home
        </Link>
      </nav>
    </header>
  );
}
