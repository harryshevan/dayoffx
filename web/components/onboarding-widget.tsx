"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslations } from "next-intl";

const EXAMPLE_TOKEN = "YOUR_MCP_TOKEN_FROM_TG_BOT";
const ONBOARDING_SEEN_KEY = "dayoffs-onboarding-seen-v1";
const ONBOARDING_QUERY_PARAM = "onboarding";
const ONBOARDING_STEP_QUERY_PARAM = "step";

type Step = {
  title: string;
  description: string;
  visual: React.ReactNode;
};

export function OnboardingWidget() {
  const t = useTranslations("onboarding");
  const [isMounted, setIsMounted] = useState(false);
  const [isVisible, setIsVisible] = useState(false);
  const [stepIndex, setStepIndex] = useState(0);
  const [isMcpCopied, setIsMcpCopied] = useState(false);
  const closeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const mcpCopyTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const botUrl = process.env.NEXT_PUBLIC_BOT_URL?.trim();
  const mcpUrl = process.env.NEXT_PUBLIC_MCP_URL?.trim() || "https://your-dayoff-host.example.com/mcp";
  const cursorConfigExample = useMemo(
    () => `{
  "mcpServers": {
    "dayoff-mcp": {
      "url": "${mcpUrl}",
      "headers": {
        "Authorization": "Bearer ${EXAMPLE_TOKEN}"
      }
    }
  }
}`,
    [mcpUrl]
  );
  const steps: Step[] = [
    {
      title: t("steps.preview.title"),
      description: t("steps.preview.description"),
      visual: (
        <div className="onb-mini-calendar" aria-hidden>
          <div className="onb-mini-calendar-title">{t("steps.preview.monthLabel")}</div>
          <div className="onb-mini-days">
            {Array.from({ length: 14 }).map((_, idx) => (
              <span key={idx} className="onb-mini-day" />
            ))}
          </div>
          <span className="onb-mini-badge">{t("steps.preview.badge")}</span>
        </div>
      )
    },
    {
      title: t("steps.statuses.title"),
      description: t("steps.statuses.description"),
      visual: (
        <div className="onb-statuses" aria-hidden>
          <div className="onb-status-chip">
            <span className="onb-dot onb-dot-pending" />
            {t("steps.statuses.pending")}
          </div>
          <div className="onb-status-chip">
            <span className="onb-dot onb-dot-approved" />
            {t("steps.statuses.approved")}
          </div>
        </div>
      )
    },
    {
      title: t("steps.admins.title"),
      description: t("steps.admins.description"),
      visual: (
        <div className="onb-profile-card" aria-hidden>
          <div className="onb-profile-avatar" />
          <div className="onb-profile-lines">
            <span />
            <span />
          </div>
          <span className="onb-mini-badge">{t("steps.admins.badge")}</span>
        </div>
      )
    },
    {
      title: t("steps.mcp.title"),
      description: t("steps.mcp.description"),
      visual: (
        <div style={{ display: "grid", gap: "0.7rem" }}>
          {botUrl ? (
            <p style={{ margin: 0, fontSize: "0.84rem", lineHeight: 1.35 }}>
              {t("steps.mcp.tokenLink.before")}{" "}
              <a href={botUrl} target="_blank" rel="noopener noreferrer">
                {t("steps.mcp.tokenLink.anchor")}
              </a>{" "}
              {t("steps.mcp.tokenLink.after")}
            </p>
          ) : null}

          <div
            style={{
              position: "relative",
              background: "#0f172a",
              color: "#e2e8f0",
              border: "1px solid #1e293b",
              borderRadius: 10,
              padding: "0.65rem"
            }}
          >
            <button
              type="button"
              className="btn"
              onClick={async () => {
                try {
                  await navigator.clipboard.writeText(cursorConfigExample);
                  setIsMcpCopied(true);
                  if (mcpCopyTimerRef.current) {
                    clearTimeout(mcpCopyTimerRef.current);
                  }
                  mcpCopyTimerRef.current = setTimeout(() => setIsMcpCopied(false), 1500);
                } catch {
                  setIsMcpCopied(false);
                }
              }}
              style={{
                position: "absolute",
                top: "0.4rem",
                right: "0.4rem",
                padding: "0.3rem 0.55rem",
                fontSize: "0.76rem"
              }}
            >
              {isMcpCopied ? t("steps.mcp.copied") : t("steps.mcp.copy")}
            </button>
            <pre
              style={{
                margin: 0,
                paddingRight: "6.8rem",
                whiteSpace: "pre-wrap",
                wordBreak: "break-word",
                fontSize: "0.74rem",
                lineHeight: 1.38,
                fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace"
              }}
            >
              {cursorConfigExample}
            </pre>
          </div>
        </div>
      )
    }
  ];

  const step = useMemo(() => steps[stepIndex], [stepIndex]);
  const isFirst = stepIndex === 0;
  const isLast = stepIndex === steps.length - 1;
  const animationMs = 220;

  const open = (initialStepIndex = 0) => {
    if (closeTimerRef.current) {
      clearTimeout(closeTimerRef.current);
      closeTimerRef.current = null;
    }
    const safeStepIndex = Math.min(Math.max(initialStepIndex, 0), steps.length - 1);
    setStepIndex(safeStepIndex);
    setIsMounted(true);
    requestAnimationFrame(() => setIsVisible(true));
  };

  const close = () => {
    setIsVisible(false);
    if (closeTimerRef.current) {
      clearTimeout(closeTimerRef.current);
    }
    closeTimerRef.current = setTimeout(() => setIsMounted(false), animationMs);
  };

  useEffect(() => {
    return () => {
      if (closeTimerRef.current) {
        clearTimeout(closeTimerRef.current);
      }
      if (mcpCopyTimerRef.current) {
        clearTimeout(mcpCopyTimerRef.current);
      }
    };
  }, []);

  useEffect(() => {
    try {
      const searchParams = new URLSearchParams(window.location.search);
      const forceOpenFromUrl = searchParams.get(ONBOARDING_QUERY_PARAM) === "1";
      const stepFromUrlRaw = searchParams.get(ONBOARDING_STEP_QUERY_PARAM);
      const parsedStep = stepFromUrlRaw ? Number.parseInt(stepFromUrlRaw, 10) : Number.NaN;
      const hasValidStepFromUrl = Number.isInteger(parsedStep);
      const stepIndexFromUrl = hasValidStepFromUrl ? parsedStep - 1 : 0;
      const hasSeenOnboarding = window.localStorage.getItem(ONBOARDING_SEEN_KEY) === "1";
      if (forceOpenFromUrl || hasValidStepFromUrl || !hasSeenOnboarding) {
        open(stepIndexFromUrl);
        window.localStorage.setItem(ONBOARDING_SEEN_KEY, "1");
      }
    } catch {
      open();
    }
  }, []);

  useEffect(() => {
    if (!isMounted) {
      return;
    }

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setIsVisible(false);
        closeTimerRef.current = setTimeout(() => setIsMounted(false), animationMs);
      }
    };

    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [isMounted]);

  return (
    <>
      <button
        type="button"
        className="onb-help-trigger"
        onClick={() => open()}
        aria-label={t("open")}
      >
        ?
      </button>

      {isMounted ? (
        <div
          className={`onb-overlay ${isVisible ? "onb-overlay-open" : "onb-overlay-closed"}`}
          role="dialog"
          aria-modal="true"
          aria-labelledby="onb-title"
          onClick={close}
        >
          <div
            className={`onb-modal card ${isVisible ? "onb-modal-open" : "onb-modal-closed"}`}
            onClick={(event) => event.stopPropagation()}
          >
            <div className="onb-topline">
              <span className="onb-progress">
                {stepIndex + 1}/{steps.length}
              </span>
              <button type="button" className="btn" onClick={close}>
                {t("close")}
              </button>
            </div>

            <h2 id="onb-title" className="onb-title">
              {step.title}
            </h2>
            <p className="onb-description">{step.description}</p>

            <div className="onb-visual-wrap">{step.visual}</div>

            <div className="onb-actions">
              <button type="button" className="btn" disabled={isFirst} onClick={() => setStepIndex((v) => v - 1)}>
                {t("back")}
              </button>
              <div style={{ display: "flex", gap: "0.5rem" }}>
                {!isLast ? (
                  <button type="button" className="btn btn-primary" onClick={() => setStepIndex((v) => v + 1)}>
                    {t("next")}
                  </button>
                ) : (
                  <button type="button" className="btn btn-primary" onClick={close}>
                    {t("gotIt")}
                  </button>
                )}
              </div>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}
