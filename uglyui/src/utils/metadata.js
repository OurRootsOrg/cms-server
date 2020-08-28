import moment from "moment";

export { getMetadataColumn, getMetadataColumnForEditing };

//Tabulator:v-data-table translation is title:text and field:value (rename "title" as "text" and "field" as "value")

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

function getMetadataColumnForEditing(pf) {
  switch (pf.type) {
    case "string":
      return {
        text: pf.name,
        value: pf.name,
        minWidth: 200,
        widthGrow: 2,
        field: pf.name,
        tooltip: pf.tooltip,
        editor: "input"
      };
    case "number":
      return {
        text: pf.name,
        value: pf.name,
        minWidth: 75,
        widthGrow: 1,
        field: pf.name,
        tooltip: pf.tooltip,
        editor: "number"
      };
    case "date":
      return {
        text: pf.name,
        value: pf.name,
        minWidth: 140,
        widthGrow: 1,
        field: pf.name,
        tooltip: pf.tooltip,
        hozAlign: "center",
        editor: dateEditor
      };
    case "boolean":
      return {
        text: pf.name,
        value: pf.name,
        minWidth: 75,
        widthGrow: 1,
        field: pf.name,
        tooltip: pf.tooltip,
        hozAlign: "center",
        formatter: "tickCross",
        editor: true
      };
  }
}

function dateEditor(cell, onRendered, success, cancel) {
  //cell - the cell component for the editable cell
  //onRendered - function to call when the editor has been rendered
  //success - function to call to pass the successfuly updated value to Tabulator
  //cancel - function to call to abort the edit and return to a normal cell

  //create and style input
  let cellValue = moment(cell.getValue(), "DD MMM YYYY").format("YYYY-MM-DD"),
    input = document.createElement("input");

  input.setAttribute("type", "date");

  input.style.padding = "4px";
  input.style.width = "100%";
  input.style.boxSizing = "border-box";

  input.value = cellValue;

  onRendered(function() {
    input.focus();
    input.style.height = "100%";
  });

  function onChange() {
    if (input.value !== cellValue) {
      success(moment(input.value, "YYYY-MM-DD").format("DD MMM YYYY"));
    } else {
      cancel();
    }
  }

  //submit new value on blur or change
  input.addEventListener("blur", onChange);

  //submit new value on enter
  input.addEventListener("keydown", function(e) {
    if (e.keyCode === 13) {
      onChange();
    }

    if (e.keyCode === 27) {
      cancel();
    }
  });

  return input;
}
