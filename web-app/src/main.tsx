import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { Providers } from "@app/providers";
import { App } from "@app/index";
import "@app/styles.css";

const savedTheme = localStorage.getItem("cd-theme");
if (savedTheme) {
  document.documentElement.setAttribute("data-theme", savedTheme);
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <Providers>
      <App />
    </Providers>
  </StrictMode>,
);
