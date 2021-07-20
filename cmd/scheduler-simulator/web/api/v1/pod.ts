import { V1Pod, V1PodList } from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const applyPod = async (req: V1Pod, id: string) => {
  const res = await instance.post<V1Pod>(`/simulators/${id}/pods`, req);
  return res.data;
};

export const listPod = async (id: string) => {
  const res = await instance.get<V1PodList>(`/simulators/${id}/pods`, {});
  return res.data;
};

export const getPod = async (name: string, id: string) => {
  const res = await instance.get<V1Pod>(`/simulators/${id}/pods/${name}`, {});
  return res.data;
};

export const deletePod = async (name: string, id: string) => {
  const res = await instance.delete(`/simulators/${id}/pods/${name}`, {});
  return res.data;
};
