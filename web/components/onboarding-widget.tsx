"use client";

import { useMemo, useState } from "react";

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
  const [isOpen, setIsOpen] = useState(false);
  const [stepIndex, setStepIndex] = useState(0);

  const step = useMemo(() => steps[stepIndex], [stepIndex]);
  const isFirst = stepIndex === 0;
  const isLast = stepIndex === steps.length - 1;

  const open = () => {
    setStepIndex(0);
    setIsOpen(true);
  };

  const close = () => setIsOpen(false);

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

      {isOpen ? (
        <div className="onb-overlay" role="dialog" aria-modal="true" aria-labelledby="onb-title">
          <div className="onb-modal card">
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
