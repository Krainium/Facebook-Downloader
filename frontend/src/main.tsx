import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { THEME } from "./brand";
import "./styles.css";

// Paint the brand palette onto :root before React mounts, so the very first
// frame is already the right colour.
for (const [k, v] of Object.entries(THEME)) document.documentElement.style.setProperty(k, v);

const el = document.getElementById("root");
if (!el) throw new Error("#root missing");

createRoot(el).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
