import type { UserConfig } from "@commitlint/types";
import { RuleConfigSeverity } from "@commitlint/types";

const Configuration: UserConfig = {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "scope-enum": [RuleConfigSeverity.Error, "always", ["books-api"]],
  },
  // ...
};

export default Configuration;
