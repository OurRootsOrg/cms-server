module.exports = {
  license: "0274a911-b8b4-46b9-9457-764421ffd41e",
  config: {
    fields: [
      {
        label: "Robot Name",
        key: "name",
        description: "The designation of the robot",
        validators: [
          {
            validate: "required_without",
            fields: ["id", "shield-color"],
            error: "must be present if no id or shield color"
          }
        ]
      },
      {
        label: "Shield Color",
        key: "shield-color",
        description: "Chromatic value",
        validators: [
          {
            validate: "regex_matches",
            regex: "^[a-zA-Z]+$",
            error: "Not alphabet only"
          }
        ]
      },
      {
        label: "Robot Helmet Style",
        key: "helmet-style",
        type: "select",
        options: [
          { value: "square", label: "Square" },
          { value: "round", label: "Round" },
          { value: "triangle", label: "Triangle" }
        ]
      },
      {
        label: "Active",
        key: "available",
        type: "checkbox"
      },
      {
        label: "Call Sign",
        key: "sign",
        alternates: ["nickname", "wave"],
        validators: [
          {
            validate: "regex_matches",
            regex: "^[a-zA-Z]{4}$",
            error: "must be 4 characters exactly"
          },
          {
            validate: "required",
            error: "required field"
          }
        ]
      },
      {
        label: "Robot ID Code",
        key: "id",
        description: "Digital identity",
        validators: [
          {
            validate: "regex_matches",
            regex: "^[0-9]*$",
            error: "must be numeric"
          },
          {
            validate: "required_without",
            fields: ["name"],
            error: "ID must be present if name is absent"
          }
        ]
      }
    ],
    type: "Robot",
    allowInvalidSubmit: false,
    managed: true,
    allowCustom: true,
    disableManualInput: false
  }
};
