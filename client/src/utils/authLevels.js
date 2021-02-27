const levels = ["Guest", "Reader", "Contributor", "Editor", "Admin"];

export { getAuthLevelName, getAuthLevelOptions };

function getAuthLevelName(level) {
  return levels[level];
}

function getAuthLevelOptions() {
  return [
    { id: 3, name: levels[3] },
    { id: 4, name: levels[4] }
  ];
}
