"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";

type Theme = "light" | "dark";

const THEME_STORAGE_KEY = "dayoffs-theme";

function getInitialTheme(): Theme {
  const documentTheme = document.documentElement.dataset.theme;
  return documentTheme === "dark" ? "dark" : "light";
}

export function ThemeToggle() {
  const t = useTranslations("theme");
  const [theme, setTheme] = useState<Theme>("light");

  useEffect(() => {
    setTheme(getInitialTheme());
  }, []);

  const nextTheme: Theme = theme === "dark" ? "light" : "dark";

  const handleClick = () => {
    document.documentElement.dataset.theme = nextTheme;
    document.documentElement.style.colorScheme = nextTheme;
    window.localStorage.setItem(THEME_STORAGE_KEY, nextTheme);
    setTheme(nextTheme);
  };

  const isDark = theme === "dark";

  return (
    <button
      type="button"
      className={`theme-switch ${isDark ? "theme-switch-on" : ""}`}
      onClick={handleClick}
      role="switch"
      aria-checked={isDark}
      aria-label={t("toggleDarkMode")}
      title={isDark ? t("darkModeOn") : t("darkModeOff")}
    >
      <span className="theme-switch-track" aria-hidden="true">
        <span className="theme-switch-thumb" />
      </span>
    </button>
  );
}
