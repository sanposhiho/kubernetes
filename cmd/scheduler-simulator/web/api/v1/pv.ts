import {
  V1PersistentVolume,
  V1PersistentVolumeList,
} from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const applyPersistentVolume = async (
  req: V1PersistentVolume,
  id: string
) => {
  const res = await instance.post<V1PersistentVolume>(
    `/simulators/${id}/persistentvolumes`,
    req
  );
  return res.data;
};

export const listPersistentVolume = async (id: string) => {
  const res = await instance.get<V1PersistentVolumeList>(
    `/simulators/${id}/persistentvolumes`,
    {}
  );
  return res.data;
};

export const deletePersistentVolume = async (name: string, id: string) => {
  const res = await instance.delete(
    `/simulators/${id}/persistentvolumes/${name}`,
    {}
  );
  return res.data;
};
