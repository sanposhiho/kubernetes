import { V1Namespace } from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const applyNamespace = async (req: V1Namespace) => {
  const res = await instance.post<V1Namespace>("/namespaces", req);
  return res.data;
};
