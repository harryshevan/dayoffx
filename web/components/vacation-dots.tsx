import type { CSSProperties } from "react";
import { VacationDot } from "@/lib/calendar";

type VacationDotsProps = {
  dots?: VacationDot[];
};

export function VacationDots({ dots }: VacationDotsProps) {
  const visibleDots = dots?.slice(0, 4) ?? [];
  const extraCount = Math.max((dots?.length ?? 0) - 4, 0);

  return (
    <>
      <span className="vac-dots">
        {visibleDots.map((dot, index) => (
          <span
            key={`${dot.colorHex}-${dot.status}-${index}`}
            className={`vac-dot ${dot.status === "pending" ? "vac-dot-pending" : ""}`}
            style={
              dot.status === "pending"
                ? ({ "--vac-dot-color": dot.colorHex } as CSSProperties)
                : { background: dot.colorHex }
            }
          />
        ))}
      </span>
      {extraCount > 0 ? <span className="vac-more">+{extraCount}</span> : null}
    </>
  );
}
