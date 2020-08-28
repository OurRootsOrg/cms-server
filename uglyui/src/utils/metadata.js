export { getMetadataColumn };

function getMetadataColumn(pf) {
  switch (pf.type) {
    case "string":
      return {
        text: pf.name,
        value: pf.name,
        tooltip: pf.tooltip,
        // headerFilter: "input",
        // sorter: "string"
      };
    case "number":
      return {
        text: pf.name,
        value: pf.name,
        tooltip: pf.tooltip,
        // headerFilter: "number",
        // sorter: "number"
      };
    case "date":
      return {
        text: pf.name,
        value: pf.name,
        // align: "center",
        tooltip: pf.tooltip,
        // headerFilter: "input",
        // sorter: "date",
        // sorterParams: {
        //   format: "DD MMM YYYY",
        //   alignEmptyValues: "top"
        // }
      };
    case "boolean":
      return {
        text: pf.name,
        value: pf.name,
        tooltip: pf.tooltip,
        // align: "center",
        // formatter: "tickCross",
        // headerFilter: "tickCross",
        // sorter: "boolean"
      };
  }
}
