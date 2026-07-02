import { join } from "node:path";
import defaultSettings from "./defaultSettings";

export default {
  alias: {
    "@root": join(__dirname, ".."),
  },
  title: "NetDisk Admin",
  layout: {
    ...defaultSettings,
  },
};
