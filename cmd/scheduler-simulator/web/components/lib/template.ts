import { V1Node, V1Pod } from "@kubernetes/client-node";
import yaml from "js-yaml";

export const podTemplate = (simulatorID: string): V1Pod => {
  if (process.env.POD_TEMPLATE !== undefined) {
    const temp = yaml.load(process.env.POD_TEMPLATE);
    temp.metadata.namespace = simulatorID;
    return temp;
  }
  return {};
};

export const nodeTemplate = (): V1Node => {
  if (process.env.NODE_TEMPLATE !== undefined) {
    return yaml.load(process.env.NODE_TEMPLATE);
  }
  return {};
};
