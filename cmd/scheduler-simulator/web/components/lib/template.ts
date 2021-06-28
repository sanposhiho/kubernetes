import { V1Node, V1Pod } from "@kubernetes/client-node";
import yaml from "js-yaml";

export const podTemplate = (): V1Pod => {
  if (process.env.POD_TEMPLATE !== undefined) {
    return yaml.load(process.env.POD_TEMPLATE);
  }
  return {};
};

export const nodeTemplate = (): V1Node => {
  if (process.env.NODE_TEMPLATE !== undefined) {
    return yaml.load(process.env.NODE_TEMPLATE);
  }
  return {};
};
