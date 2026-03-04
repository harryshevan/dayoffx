"use client";

import { useEffect, useMemo, useState } from "react";
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
    vacations.reduce<Map<string, { name: string; color: string }>>((acc, item) => {
      if (!acc.has(item.memberId)) {
        acc.set(item.memberId, { name: item.displayName, color: item.colorHex });
      }
      return acc;
    }, new Map()).values()
  );

  return (
    <main className="grid">
      <section className="card">
        <div style={{ display: "flex", justifyContent: "space-between", gap: "0.75rem", flexWrap: "wrap" }}>
          <h1 style={{ margin: 0 }}>Годовой календарь отпусков</h1>
          <label>
            Год{" "}
            <select value={year} onChange={(event) => setYear(Number(event.target.value))}>
              {yearOptions.map((optionYear) => (
                <option key={optionYear} value={optionYear}>
                  {optionYear}
                </option>
              ))}
            </select>
          </label>
        </div>
      </section>

      {legendByMember.length > 0 ? (
        <section className="legend">
          {legendByMember.map((item) => (
            <div key={item.name} className="legend-item">
              <span className="dot" style={{ background: item.color }} />
              {item.name}
            </div>
          ))}
        </section>
      ) : null}

      {error ? <section className="card">Ошибка: {error}</section> : null}
      {loading ? <section className="card">Загружаем календарь...</section> : null}
      {!loading && !error ? <YearCalendar year={year} vacations={vacations} /> : null}
    </main>
  );
}
