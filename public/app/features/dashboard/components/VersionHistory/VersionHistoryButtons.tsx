import React from 'react';

import { HorizontalGroup, Tooltip, Button } from '@grafana/ui';

type VersionsButtonsType = {
  hasMore: boolean;
  canCompare: boolean;
  getVersions: (append: boolean) => void;
  getDiff: () => void;
  isLastPage: boolean;
};
export const VersionsHistoryButtons: React.FC<VersionsButtonsType> = ({
  hasMore,
  canCompare,
  getVersions,
  getDiff,
  isLastPage,
}) => (
  <HorizontalGroup>
    {hasMore && (
      <Button type="button" onClick={() => getVersions(true)} variant="secondary" disabled={isLastPage}>
        Show more versions
      </Button>
    )}
    <Tooltip content="Select two versions to start comparing" placement="bottom">
      <Button type="button" disabled={!canCompare} onClick={getDiff} icon="code-branch">
        Compare versions
      </Button>
    </Tooltip>
  </HorizontalGroup>
);
