export const objectToTreeViewData = (ent: Object | null): Array<object> => {
  if (ent == null) {
    return [];
  }

  const data = [];
  for (const [key, value] of Object.entries(ent)) {
    if (typeof value == "object") {
      data.push({
        id: key,
        name: key,
        children: objectToTreeViewData(value),
      });
    } else if (Array.isArray(value)) {
      data.push({
        id: key,
        name: key,
        children: objectToTreeViewData(value),
      });
    } else {
      data.push({
        id: key,
        name: key + ": " + value,
      });
    }
  }
  return data;
};
