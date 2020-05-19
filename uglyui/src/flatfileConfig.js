module.exports = {
  license: "0274a911-b8b4-46b9-9457-764421ffd41e",
  config: {
    fields: [
      {
        label: "Given Name",
        key: "given",
        description: "The person's given name"
      },
      {
        label: "Surname",
        key: "surname",
        description: "The person's surname",
        validators: [
          {
            validate: "required",
            error: "required field"
          }
        ]
      }
    ],
    type: "Person",
    allowInvalidSubmit: true,
    managed: true,
    allowCustom: false,
    disableManualInput: true
  }
};
