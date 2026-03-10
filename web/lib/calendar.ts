import { DayOffOverride, Vacation } from "./types";

export type VacationDot = {
  colorHex: string;
  status: Vacation["status"];
};

export type DayVacation = {
  memberId: string;
  displayName: string;
  colorHex: string;
  status: Vacation["status"];
  fromDate: string;
  toDate: string;
};

export type DayCell = {
  day: number;
  date: Date;
  inCurrentMonth: boolean;
  isDayOff: boolean;
  vacationDots?: VacationDot[];
  vacationNames?: string[];
  vacationMemberIds?: string[];
  vacations?: DayVacation[];
};

const WEEK_START_MONDAY = 1;

function toIsoDateString(value: Date): string {
  const year = value.getUTCFullYear();
  const month = `${value.getUTCMonth() + 1}`.padStart(2, "0");
  const day = `${value.getUTCDate()}`.padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function buildMonthGrid(
  year: number,
  monthIndex: number,
  vacations: Vacation[],
  dayOffOverrides: DayOffOverride[]
): DayCell[] {
  const dayOffLookup = new Map<string, boolean>();
  for (const item of dayOffOverrides) {
    dayOffLookup.set(item.date, item.isDayOff);
  }
  const firstDayOfMonth = new Date(Date.UTC(year, monthIndex, 1));
  const firstWeekday = (firstDayOfMonth.getUTCDay() + 7 - WEEK_START_MONDAY) % 7;
  const gridStart = new Date(firstDayOfMonth);
  gridStart.setUTCDate(gridStart.getUTCDate() - firstWeekday);

  return Array.from({ length: 42 }, (_, index) => {
    const date = new Date(gridStart);
    date.setUTCDate(gridStart.getUTCDate() + index);
    const dateString = toIsoDateString(date);
    const overrideValue = dayOffLookup.get(dateString);
    const isWeekend = date.getUTCDay() === 0 || date.getUTCDay() === 6;
    const isDayOff = overrideValue ?? isWeekend;

    const vacationsForDay = vacations.filter((item) => dateString >= item.fromDate && dateString <= item.toDate);
    return {
      day: date.getUTCDate(),
      date,
      inCurrentMonth: date.getUTCMonth() === monthIndex,
      isDayOff,
      vacationDots:
        vacationsForDay.length > 0
          ? vacationsForDay.map((item) => ({ colorHex: item.colorHex, status: item.status }))
          : undefined,
      vacationNames: vacationsForDay.length > 0 ? vacationsForDay.map((item) => item.displayName) : undefined,
      vacationMemberIds: vacationsForDay.length > 0 ? vacationsForDay.map((item) => item.memberId) : undefined,
      vacations:
        vacationsForDay.length > 0
          ? vacationsForDay.map((item) => ({
              memberId: item.memberId,
              displayName: item.displayName,
              colorHex: item.colorHex,
              status: item.status,
              fromDate: item.fromDate,
              toDate: item.toDate
            }))
          : undefined
    };
  });
}
