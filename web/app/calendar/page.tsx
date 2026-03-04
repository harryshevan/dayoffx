"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { getVacations } from "@/lib/api";
import { Vacation } from "@/lib/types";
import { YearCalendar } from "@/components/year-calendar";

function getCurrentYear(): number {
  return new Date().getUTCFullYear();
}

export default function CalendarPage() {
  const currentYear = getCurrentYear();
  const [year, setYear] = useState(currentYear);
  const [vacations, setVacations] = useState<Vacation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [hoveredMemberId, setHoveredMemberId] = useState<string | null>(null);
  const [selectedMemberId, setSelectedMemberId] = useState<string | null>(null);
  const legendRef = useRef<HTMLElement | null>(null);

  const yearOptions = useMemo(
    () => Array.from({ length: 5 }, (_, index) => currentYear - 2 + index),
    [currentYear]
  );

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    getVacations(year)
      .then((data) => {
        if (!cancelled) {
          setVacations(data);
        }
      })
      .catch((requestError) => {
        if (!cancelled) {
          setError(requestError instanceof Error ? requestError.message : "Failed to load vacations");
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [year]);

  const legendByMember = Array.from(
    vacations.reduce<Map<string, { memberId: string; name: string; color: string }>>((acc, item) => {
      if (!acc.has(item.memberId)) {
        acc.set(item.memberId, { memberId: item.memberId, name: item.displayName, color: item.colorHex });
      }
      return acc;
    }, new Map()).values()
  );
  const highlightedMemberId = selectedMemberId ?? hoveredMemberId;

  useEffect(() => {
    const handlePointerDown = (event: PointerEvent) => {
      if (!selectedMemberId) {
        return;
      }
      if (legendRef.current?.contains(event.target as Node)) {
        return;
      }
      setSelectedMemberId(null);
      setHoveredMemberId(null);
    };

    document.addEventListener("pointerdown", handlePointerDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
    };
  }, [selectedMemberId]);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key !== "Escape") {
        return;
      }
      if (!selectedMemberId && !hoveredMemberId) {
        return;
      }
      setSelectedMemberId(null);
      setHoveredMemberId(null);
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [selectedMemberId, hoveredMemberId]);

  return (
    <main className="grid">
      {legendByMember.length > 0 ? (
        <section ref={legendRef} className="legend">
          {legendByMember.map((item) => (
            <button
              key={item.memberId}
              type="button"
              className={`legend-item ${highlightedMemberId === item.memberId ? "legend-item-active" : ""}`}
              onMouseEnter={() => setHoveredMemberId(item.memberId)}
              onMouseLeave={() => setHoveredMemberId(null)}
              onClick={() => setSelectedMemberId((previous) => (previous === item.memberId ? null : item.memberId))}
              aria-pressed={selectedMemberId === item.memberId}
              aria-label={`${item.name}: ${selectedMemberId === item.memberId ? "disable" : "enable"} filter`}
            >
              <span className="dot" style={{ background: item.color }} />
              {item.name}
            </button>
          ))}
        </section>
      ) : null}

      {error ? <section className="card">Error: {error}</section> : null}
      {loading ? (
        <section className="calendar-skeleton" aria-label="Loading calendar" aria-busy="true">
          {Array.from({ length: 6 }, (_, monthIndex) => (
            <div key={monthIndex} className="skeleton-month">
              <div className="skeleton-title" />
              <div className="skeleton-grid">
                {Array.from({ length: 35 }, (_, dayIndex) => (
                  <div key={dayIndex} className="skeleton-cell" />
                ))}
              </div>
            </div>
          ))}
        </section>
      ) : null}
      {!loading && !error ? (
        <>
          <section className="calendar-controls">
            <label className="calendar-year-select">
              Year{" "}
              <select value={year} onChange={(event) => setYear(Number(event.target.value))}>
                {yearOptions.map((optionYear) => (
                  <option key={optionYear} value={optionYear}>
                    {optionYear}
                  </option>
                ))}
              </select>
            </label>
          </section>
          <YearCalendar year={year} vacations={vacations} highlightedMemberId={highlightedMemberId} />
        </>
      ) : null}
    </main>
  );
}
