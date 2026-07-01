import { Flex, Tooltip } from "antd";
import { CopyOutlined } from "@ant-design/icons";
import { useState } from "react";
import { useTranslation } from "react-i18next";

type Props = {
  value: string | number;
  /** Render children instead of plain value text */
  children?: React.ReactNode;
};

export default function CopyCell({ value, children }: Props) {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(String(value));
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // clipboard not available
    }
  };

  return (
    <Flex gap={16}>
      <span
        style={{
          overflow: "hidden",
          textOverflow: "ellipsis",
          whiteSpace: "nowrap",
          flex: 1,
        }}
      >
        {children ?? String(value)}
      </span>
      <Tooltip title={copied ? t("common.copied") : t("common.copy")}>
        <CopyOutlined
          onClick={handleCopy}
          style={{
            marginLeft: 6,
            cursor: "pointer",
            fontSize: 13,
            color: copied ? "#52c41a" : "#8c8c8c",
            transition: "color 0.2s",
          }}
        />
      </Tooltip>
    </Flex>
  );
}
