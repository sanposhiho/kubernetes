import {
  V1Node,
  V1PersistentVolume,
  V1PersistentVolumeClaim,
  V1Pod,
  V1StorageClass,
} from "@kubernetes/client-node";
import yaml from "js-yaml";

export const podTemplate = (simulatorID: string): V1Pod => {
  if (process.env.POD_TEMPLATE) {
    const temp = yaml.load(process.env.POD_TEMPLATE);
    temp.metadata.namespace = simulatorID;
    return temp;
  }
  return {};
};

export const nodeTemplate = (): V1Node => {
  if (process.env.NODE_TEMPLATE) {
    return yaml.load(process.env.NODE_TEMPLATE);
  }
  return {};
};

export const pvTemplate = (): V1PersistentVolume => {
  if (process.env.PV_TEMPLATE) {
    return yaml.load(process.env.PV_TEMPLATE);
  }
  return {};
};

export const pvcTemplate = (): V1PersistentVolumeClaim => {
  if (process.env.PVC_TEMPLATE) {
    return yaml.load(process.env.PVC_TEMPLATE);
  }
  return {};
};

export const storageclassTemplate = (): V1StorageClass => {
  if (process.env.SC_TEMPLATE) {
    return yaml.load(process.env.SC_TEMPLATE);
  }
  return { provisioner: "" };
};
