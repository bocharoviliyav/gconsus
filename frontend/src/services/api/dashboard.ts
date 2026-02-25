import { apiClient } from "./client";
import type { DashboardStats } from "../../types";

export const getDashboard = async (
  start: string,
  end: string,
): Promise<DashboardStats> => {
  const response = await apiClient.get("/analytics/dashboard", {
    params: { start, end },
  });
  return response.data;
};
