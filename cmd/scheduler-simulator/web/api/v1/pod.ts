import { V1Pod, V1PodList } from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const createPod = async (req: V1Pod) => {
  const res = await instance.post<V1Pod>("/pods", {
    metadata: {
      name: req.metadata?.name,
    },
  });
  return res.data;
};

export const listPod = async (req: ListPodReq) => {
  const res = await instance.get<V1PodList>("/pods", {});
  return res.data;
};
interface ListPodReq {}
