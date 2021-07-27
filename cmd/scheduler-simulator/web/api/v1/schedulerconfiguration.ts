import { instance } from "@/api/v1/index";
import { SchedulerConfiguration } from "./types";

export const applySchedulerConfiguration = async (
  req: SchedulerConfiguration,
  id: string
) => {
  const res = await instance.post<SchedulerConfiguration>(
    `/simulators/${id}/schedulerconfiguration`,
    req
  );
  return res.data;
};

export const getSchedulerConfiguration = async (id: string) => {
  const res = await instance.get<SchedulerConfiguration>(
    `/simulators/${id}/schedulerconfiguration`
  );
  return res.data;
};
