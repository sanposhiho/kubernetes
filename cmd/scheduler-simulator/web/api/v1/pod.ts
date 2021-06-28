import { V1Pod, V1PodList } from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const applyPod = async (req: V1Pod) => {
  const res = await instance.post<V1Pod>("/pods", req);
  return res.data;
};

export const listPod = async () => {
  const res = await instance.get<V1PodList>("/pods", {});
  return res.data;
};

export const deletePod = async (name: string) => {
  const res = await instance.delete(`/pods/${name}`, {});
  return res.data;
};
