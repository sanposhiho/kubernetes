import { V1StorageClass, V1StorageClassList } from "@kubernetes/client-node";
import { instance } from "@/api/v1/index";

export const applyStorageClass = async (
  req: V1StorageClass,
  id: string,
  onError: (msg: string) => void
) => {
  try {
    const res = await instance.post<V1StorageClass>(
      `/simulators/${id}/storageclasses`,
      req
    );
    return res.data;
  } catch (e) {
    onError(e);
  }
};

export const listStorageClass = async (id: string) => {
  const res = await instance.get<V1StorageClassList>(
    `/simulators/${id}/storageclasses`,
    {}
  );
  return res.data;
};

export const getStorageClass = async (name: string, id: string) => {
  const res = await instance.get<V1StorageClass>(
    `/simulators/${id}/storageclasses/${name}`,
    {}
  );
  return res.data;
};

export const deleteStorageClass = async (name: string, id: string) => {
  const res = await instance.delete(
    `/simulators/${id}/storageclasses/${name}`,
    {}
  );
  return res.data;
};
