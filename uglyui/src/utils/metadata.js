export { getMetadataColumn };

function getMetadataColumn(pf) {
  switch (pf.type) {
    case "string":
      return {
        text: pf.name,
        value: pf.name,
        tooltip: pf.tooltip
      };
    case "number":
      return {
        text: pf.name,
        value: pf.name,
        tooltip: pf.tooltip
      };
    case "date":
      return {
        text: pf.name,
        value: pf.name,
        tooltip: pf.tooltip
      };
    case "boolean":
      return {
        text: pf.name,
        value: pf.name,
        tooltip: pf.tooltip
      };
  }
}
