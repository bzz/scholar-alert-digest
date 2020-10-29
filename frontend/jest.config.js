module.exports = {
  moduleNameMapper: {
    "\\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2)$": "./jest.mock.js",
  },
  testPathIgnorePatterns: ["<rootDir>/node_modules/"],
  transform: {
    "^.+\\.jsx?$": "babel-jest",
    "\\.(css|less|scss|sass)$": "jest-transform-css",
  },
}
