"use client";

import { useEffect, useMemo, useRef, useState } from "react";

type Step = {
  title: string;
  description: string;
  visual: React.ReactNode;
};

const steps: Step[] = [
  {
    title: "Calendar is preview only",
    description: "Web calendar is read-only preview. Requests are managed via MCP (and later via Telegram bot).",
    visual: (
      <div className="onb-mini-calendar" aria-hidden>
        <div className="onb-mini-calendar-title">March 2026</div>
        <div className="onb-mini-days">
          {Array.from({ length: 14 }).map((_, idx) => (
            <span key={idx} className="onb-mini-day" />
          ))}
        </div>
        <span className="onb-mini-badge">Preview</span>
      </div>
    )
  },
  {
    title: "Two request statuses",
    description: "Pending request is a half-circle dot. Approved request is a filled circle dot.",
    visual: (
      <div className="onb-statuses" aria-hidden>
        <div className="onb-status-chip">
          <span className="onb-dot onb-dot-pending" />
          Pending
        </div>
        <div className="onb-status-chip">
          <span className="onb-dot onb-dot-approved" />
          Approved
        </div>
      </div>
    )
  },
  {
    title: "Admins approve requests",
    description: "Admins review and approve requests. Your profile name and color can be changed via MCP config.",
    visual: (
      <div className="onb-profile-card" aria-hidden>
        <div className="onb-profile-avatar" />
        <div className="onb-profile-lines">
          <span />
          <span />
        </div>
        <span className="onb-mini-badge">MCP config</span>
      </div>
    )
  }
];

export function OnboardingWidget() {
  const [isMounted, setIsMounted] = useState(false);
  const [isVisible, setIsVisible] = useState(false);
  const [stepIndex, setStepIndex] = useState(0);
  const closeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const step = useMemo(() => steps[stepIndex], [stepIndex]);
  const isFirst = stepIndex === 0;
  const isLast = stepIndex === steps.length - 1;
  const animationMs = 220;

  useEffect(() => {
    return () => {
      if (closeTimerRef.current) {
        clearTimeout(closeTimerRef.current);
      }
    };
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

  const open = () => {
    if (closeTimerRef.current) {
      clearTimeout(closeTimerRef.current);
      closeTimerRef.current = null;
    }
    setStepIndex(0);
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

  return (
    <>
      <button
        type="button"
        className="onb-help-trigger"
        onClick={open}
        aria-label="Open onboarding"
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
                Close
              </button>
            </div>

            <h2 id="onb-title" className="onb-title">
              {step.title}
            </h2>
            <p className="onb-description">{step.description}</p>

            <div className="onb-visual-wrap">{step.visual}</div>

            <div className="onb-actions">
              <button type="button" className="btn" disabled={isFirst} onClick={() => setStepIndex((v) => v - 1)}>
                Back
              </button>
              <div style={{ display: "flex", gap: "0.5rem" }}>
                {!isLast ? (
                  <button type="button" className="btn btn-primary" onClick={() => setStepIndex((v) => v + 1)}>
                    Next
                  </button>
                ) : (
                  <button type="button" className="btn btn-primary" onClick={close}>
                    Got it
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
