module.exports = {
  env: {
    browser: true,
    es2020: true,
  },
  extends: [
    "plugin:react/recommended",
    "airbnb",
  ],
  parserOptions: {
    ecmaFeatures: {
      jsx: true,
    },
    ecmaVersion: 11,
    sourceType: "module",
  },
  plugins: [
    "react",
  ],
  rules: {
    semi: 0,
    "no-shadow": 0,
    "arrow-parens": 0,
    "object-curly-spacing": 0,
    "object-curly-newline": 0,
    "no-confusing-arrow": 0,
    "react/jsx-filename-extension": 0,
    "react/jsx-one-expression-per-line": 0,
    "operator-linebreak": ["error", "after"],
    "implicit-arrow-linebreak": 0,
    "function-paren-newline": 0,
    "operator-assignment": 0,
    "import/no-extraneous-dependencies": ["error", {devDependencies: true}],
    "import/prefer-default-export": 0,
    "no-unused-vars": ["error", {
      vars: "all",
      args: "after-used",
      ignoreRestSiblings: false,
      varsIgnorePattern: "_",
      argsIgnorePattern: "_",
    }],
    quotes: ["error", "double"],
  },
  settings: {
    "import/resolver": {
      "babel-module": {},
    },
  },
}
