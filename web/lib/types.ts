export type Goal = "cursor" | "claude_desktop" | "other";

export type MemberRole = "member" | "admin";

export type MemberConnection = {
  memberId: string;
  displayName: string;
  colorHex: string;
  role: MemberRole;
  mcpToken: string;
  mcpServerUrl?: string;
};

export type VacationStatus = "pending" | "approved";

export type Vacation = {
  id: string;
  memberId: string;
  displayName: string;
  colorHex: string;
  fromDate: string;
  toDate: string;
  reason: string;
  status: VacationStatus;
};

export type DayOffOverride = {
  date: string;
  isDayOff: boolean;
  reason: string;
};
