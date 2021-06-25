import { V1Node, V1NodeList, V1Pod, V1PodList } from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const applyNode = async (req: V1Node) => {
  const res = await instance.post<V1Node>("/nodes", {
    metadata: {
      name: req.metadata?.name,
      labels: req.metadata?.labels,
    },
  });
  return res.data;
};

export const listNode = async (req: ListNodeReq) => {
  const res = await instance.get<V1NodeList>("/nodes", {});
  return res.data;
};

interface ListNodeReq {}
