import { V1Namespace } from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const createNamespace = async () => {
  const res = await instance.post<V1Namespace>("/namespaces");
  return res.data;
};

export const getNamespace = async (name: string) => {
  const res = await instance.get<V1Namespace>(`/namespaces/${name}`);
  return res.data;
};
