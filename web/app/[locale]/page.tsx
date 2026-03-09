"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import { getVacations } from "@/lib/api";
import { Vacation } from "@/lib/types";
import { YearCalendar } from "@/components/year-calendar";

function getCurrentYear(): number {
  return new Date().getUTCFullYear();
}

export default function CalendarPage() {
  const t = useTranslations("calendar");
  const currentYear = getCurrentYear();
  const [year, setYear] = useState(currentYear);
  const [vacations, setVacations] = useState<Vacation[]>([]);
  const [isInitialLoading, setIsInitialLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hoveredMemberId, setHoveredMemberId] = useState<string | null>(null);
  const [selectedMemberIds, setSelectedMemberIds] = useState<string[]>([]);
  const legendRef = useRef<HTMLElement | null>(null);
  const requestIdRef = useRef(0);
  const hasLoadedOnceRef = useRef(false);

  const yearOptions = useMemo(
    () => Array.from({ length: 5 }, (_, index) => currentYear - 2 + index),
    [currentYear]
  );

  useEffect(() => {
    const requestId = requestIdRef.current + 1;
    requestIdRef.current = requestId;
    const isFirstLoad = !hasLoadedOnceRef.current;

    if (isFirstLoad) {
      setIsInitialLoading(true);
    } else {
      setIsRefreshing(true);
    }
    setError(null);

    getVacations(year)
      .then((data) => {
        if (requestIdRef.current === requestId) {
          setVacations(data);
          hasLoadedOnceRef.current = true;
        }
      })
      .catch((requestError) => {
        if (requestIdRef.current === requestId) {
          setError(requestError instanceof Error ? requestError.message : t("failedToLoadVacations"));
        }
      })
      .finally(() => {
        if (requestIdRef.current === requestId) {
          if (isFirstLoad) {
            setIsInitialLoading(false);
          }
          setIsRefreshing(false);
        }
      });
  }, [year, t]);

  const legendByMember = Array.from(
    vacations.reduce<Map<string, { memberId: string; name: string; color: string }>>((acc, item) => {
      if (!acc.has(item.memberId)) {
        acc.set(item.memberId, { memberId: item.memberId, name: item.displayName, color: item.colorHex });
      }
      return acc;
    }, new Map()).values()
  );
  const highlightedMemberIds =
    selectedMemberIds.length > 0 ? selectedMemberIds : hoveredMemberId ? [hoveredMemberId] : [];
  const isBusy = isInitialLoading || isRefreshing;

  useEffect(() => {
    const handlePointerDown = (event: PointerEvent) => {
      if (selectedMemberIds.length === 0) {
        return;
      }
      if (legendRef.current?.contains(event.target as Node)) {
        return;
      }
      setSelectedMemberIds([]);
      setHoveredMemberId(null);
    };

    document.addEventListener("pointerdown", handlePointerDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
    };
  }, [selectedMemberIds]);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key !== "Escape") {
        return;
      }
      if (selectedMemberIds.length === 0 && !hoveredMemberId) {
        return;
      }
      setSelectedMemberIds([]);
      setHoveredMemberId(null);
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [selectedMemberIds, hoveredMemberId]);

  return (
    <main className="grid">
      <section className="calendar-controls">
        <label className="calendar-year-select">
          {t("year")}{" "}
          <select value={year} onChange={(event) => setYear(Number(event.target.value))} aria-busy={isBusy}>
            {yearOptions.map((optionYear) => (
              <option key={optionYear} value={optionYear}>
                {optionYear}
              </option>
            ))}
          </select>
        </label>
      </section>

      {!isInitialLoading ? (
        <section ref={legendRef} className={`legend ${legendByMember.length === 0 ? "legend-empty" : ""}`}>
          {legendByMember.map((item) => (
            <button
              key={item.memberId}
              type="button"
              className={`legend-item ${selectedMemberIds.includes(item.memberId) ? "legend-item-active" : ""}`}
              onMouseEnter={() => {
                if (selectedMemberIds.length === 0) {
                  setHoveredMemberId(item.memberId);
                }
              }}
              onMouseLeave={() => setHoveredMemberId(null)}
              onClick={() =>
                setSelectedMemberIds((previous) =>
                  previous.includes(item.memberId)
                    ? previous.filter((memberId) => memberId !== item.memberId)
                    : [...previous, item.memberId]
                )
              }
              aria-pressed={selectedMemberIds.includes(item.memberId)}
              aria-label={t(
                selectedMemberIds.includes(item.memberId) ? "legendRemoveFromFilter" : "legendAddToFilter",
                { name: item.name }
              )}
            >
              <span className="dot" style={{ background: item.color }} />
              {item.name}
            </button>
          ))}
        </section>
      ) : null}

      {error ? <section className="card">{t("error")}: {error}</section> : null}
      <section className="calendar-stage" aria-busy={isBusy}>
        {isInitialLoading ? (
          <section className="calendar-skeleton" aria-label={t("loadingCalendar")}>
            {Array.from({ length: 12 }, (_, monthIndex) => (
              <div key={monthIndex} className="skeleton-month">
                <div className="skeleton-title" />
                <div className="skeleton-weekday-grid" aria-hidden="true">
                  {Array.from({ length: 7 }, (_, weekdayIndex) => (
                    <div key={weekdayIndex} className="skeleton-weekday" />
                  ))}
                </div>
                <div className="skeleton-grid">
                  {Array.from({ length: 42 }, (_, dayIndex) => (
                    <div key={dayIndex} className="skeleton-cell" />
                  ))}
                </div>
              </div>
            ))}
          </section>
        ) : (
          <YearCalendar year={year} vacations={vacations} highlightedMemberIds={highlightedMemberIds} />
        )}
        {isRefreshing ? (
          <div className="calendar-refresh-overlay" role="status" aria-live="polite">
            <span className="calendar-refresh-chip">{t("loadingYear", { year })}</span>
          </div>
        ) : null}
      </section>
      {!isInitialLoading && !error && vacations.length === 0 ? (
        <section className="card">{t("noVacationsForYear", { year })}</section>
      ) : null}
    </main>
  );
}
