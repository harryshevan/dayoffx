import { Vacation } from "./types";

export type DayCell = {
  day: number;
  date: Date;
  inCurrentMonth: boolean;
  vacationColor?: string;
  vacationName?: string;
};

const WEEK_START_MONDAY = 1;

function toIsoDateString(value: Date): string {
  const year = value.getUTCFullYear();
  const month = `${value.getUTCMonth() + 1}`.padStart(2, "0");
  const day = `${value.getUTCDate()}`.padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function buildMonthGrid(year: number, monthIndex: number, vacations: Vacation[]): DayCell[] {
  const firstDayOfMonth = new Date(Date.UTC(year, monthIndex, 1));
  const firstWeekday = (firstDayOfMonth.getUTCDay() + 7 - WEEK_START_MONDAY) % 7;
  const gridStart = new Date(firstDayOfMonth);
  gridStart.setUTCDate(gridStart.getUTCDate() - firstWeekday);

  return Array.from({ length: 42 }, (_, index) => {
    const date = new Date(gridStart);
    date.setUTCDate(gridStart.getUTCDate() + index);
    const dateString = toIsoDateString(date);

    const vacation = vacations.find((item) => dateString >= item.fromDate && dateString <= item.toDate);
    return {
      day: date.getUTCDate(),
      date,
      inCurrentMonth: date.getUTCMonth() === monthIndex,
      vacationColor: vacation?.colorHex,
      vacationName: vacation?.displayName
    };
  });
}
