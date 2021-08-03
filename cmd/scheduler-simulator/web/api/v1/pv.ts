import {
  V1PersistentVolume,
  V1PersistentVolumeList,
} from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const applyPersistentVolume = async (
  req: V1PersistentVolume,
  id: string,
  onError: (msg: string) => void
) => {
  try {
    const res = await instance.post<V1PersistentVolume>(
      `/simulators/${id}/persistentvolumes`,
      req
    );
    return res.data;
  } catch (e) {
    onError(e);
  }
};

export const listPersistentVolume = async (id: string) => {
  const res = await instance.get<V1PersistentVolumeList>(
    `/simulators/${id}/persistentvolumes`,
    {}
  );
  return res.data;
};

export const getPersistentVolume = async (name: string, id: string) => {
  const res = await instance.get<V1PersistentVolume>(
    `/simulators/${id}/persistentvolumes/${name}`,
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
