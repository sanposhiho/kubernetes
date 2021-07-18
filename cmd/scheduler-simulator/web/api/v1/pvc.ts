import {
  V1PersistentVolumeClaim,
  V1PersistentVolumeClaimList,
  V1Pod,
  V1PodList,
} from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const applyPersistentVolumeClaim = async (
  req: V1PersistentVolumeClaim,
  id: string
) => {
  const res = await instance.post<V1PersistentVolumeClaim>(
    `/simulators/${id}/persistentvolumeclaims`,
    req
  );
  return res.data;
};

export const listPersistentVolumeClaim = async (id: string) => {
  const res = await instance.get<V1PersistentVolumeClaimList>(
    `/simulators/${id}/persistentvolumeclaims`,
    {}
  );
  return res.data;
};

export const getPersistentVolumeClaim = async (name: string, id: string) => {
  const res = await instance.get<V1PersistentVolumeClaim>(
    `/simulators/${id}/persistentvolumeclaims/${name}`,
    {}
  );
  return res.data;
};

export const deletePersistentVolumeClaim = async (name: string, id: string) => {
  const res = await instance.delete(
    `/simulators/${id}/persistentvolumeclaims/${name}`,
    {}
  );
  return res.data;
};
