import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { RootState } from '@/store';

interface WorkflowDesignerState {
  selectedNodeId: string | null;
  previewMode: boolean;
  definitionId: string | null;
  definitionName: string;
  definitionKey: string;
  definitionStatus: number;
  dirty: boolean;
}

const initialState: WorkflowDesignerState = {
  selectedNodeId: null,
  previewMode: false,
  definitionId: null,
  definitionName: '',
  definitionKey: '',
  definitionStatus: 0,
  dirty: false,
};

const workflowSlice = createSlice({
  name: 'workflow',
  initialState,
  reducers: {
    setSelectedNodeId(state, action: PayloadAction<string | null>) {
      state.selectedNodeId = action.payload;
    },
    setPreviewMode(state, action: PayloadAction<boolean>) {
      state.previewMode = action.payload;
    },
    setDefinitionMeta(
      state,
      action: PayloadAction<{
        id: string | null;
        name: string;
        key: string;
        status: number;
      }>,
    ) {
      state.definitionId = action.payload.id;
      state.definitionName = action.payload.name;
      state.definitionKey = action.payload.key;
      state.definitionStatus = action.payload.status;
    },
    setDirty(state, action: PayloadAction<boolean>) {
      state.dirty = action.payload;
    },
    resetDesigner() {
      return initialState;
    },
  },
});

export const {
  setSelectedNodeId,
  setPreviewMode,
  setDefinitionMeta,
  setDirty,
  resetDesigner,
} = workflowSlice.actions;

export const selectSelectedNodeId = (state: RootState) => state.workflow.selectedNodeId;
export const selectPreviewMode = (state: RootState) => state.workflow.previewMode;
export const selectDefinitionId = (state: RootState) => state.workflow.definitionId;
export const selectDefinitionName = (state: RootState) => state.workflow.definitionName;
export const selectDefinitionKey = (state: RootState) => state.workflow.definitionKey;
export const selectDesignerDirty = (state: RootState) => state.workflow.dirty;

export default workflowSlice.reducer;
