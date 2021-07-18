import { V1Node, V1NodeList, V1Pod, V1PodList } from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const applyNode = async (req: V1Node, id: string) => {
  const res = await instance.post<V1Node>(`/simulators/${id}/nodes`, req);
  return res.data;
};

export const listNode = async (id: string) => {
  const res = await instance.get<V1NodeList>(`/simulators/${id}/nodes`, {});
  return res.data;
};

export const getNode = async (name: string, id: string) => {
  const res = await instance.get<V1Node>(`/simulators/${id}/nodes/${name}`, {});
  return res.data;
};

export const deleteNode = async (name: string, id: string) => {
  const res = await instance.delete(`/simulators/${id}/nodes/${name}`, {});
  return res.data;
};
