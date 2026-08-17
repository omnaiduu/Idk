import { useCallback, useEffect, useMemo, useState } from "react";
import * as api from "@/api/client";
import { BenchSkeleton } from "@/components/BenchSkeleton";
import { BottomPanel } from "@/components/BottomPanel";
import { CodeEditor } from "@/components/CodeEditor";
import { ErrorSlideOver } from "@/components/ErrorSlideOver";
import { GateCountDialog } from "@/components/GateCountDialog";
import { LeftRail } from "@/components/LeftRail";
import { McpDrawer } from "@/components/McpDrawer";
import { OnboardingOverlay } from "@/components/OnboardingOverlay";
import { RightRail } from "@/components/RightRail";
import { TopBar } from "@/components/TopBar";
import { TooltipProvider } from "@/components/ui/tooltip";
import { useBench } from "@/hooks/use-bench";
import { BLOCKS, ONBOARDING_KEY, type BlockId } from "@/lib/utils";
import type { ApiError, ServosState, WavesState } from "@/types";

const DEFAULT_C = `#include <stdint.h>

/* hello-gpu template — edit and press Run */
void main(void) {
}
`;

const DEFAULT_V = `// Select a block on the left to edit HDL files.
`;

export default function App() {
  const bench = useBench();
  const [selectedBlock, setSelectedBlock] = useState<BlockId>("cpu");
  const [editorPath, setEditorPath] = useState("firmware/main.c");
  const [editorContent, setEditorContent] = useState(DEFAULT_C);
  const [editorDirty, setEditorDirty] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);
  const [hasRun, setHasRun] = useState(false);
  const [uart, setUart] = useState("");
  const [leds, setLeds] = useState(0);
  const [servos, setServos] = useState<ServosState>({ duty: [0, 0, 0, 0] });
  const [waves, setWaves] = useState<WavesState>({
    clk: [],
    gpu_busy: [],
    led0: [],
    pwm0: [],
  });
  const [buttonDown, setButtonDown] = useState(false);
  const [gateOpen, setGateOpen] = useState(false);
  const [gateLoading, setGateLoading] = useState(false);
  const [gateCells, setGateCells] = useState<number>();
  const [gateWires, setGateWires] = useState<number>();
  const [gateError, setGateError] = useState<string>();
  const [onboardingOpen, setOnboardingOpen] = useState(false);

  const blockFileMap = useMemo(
    () => Object.fromEntries(BLOCKS.map((b) => [b.id, b.file])) as Record<BlockId, string>,
    [],
  );

  useEffect(() => {
    const seen = localStorage.getItem(ONBOARDING_KEY);
    if (!seen) setOnboardingOpen(true);
  }, []);

  const loadFile = useCallback(async (path: string) => {
    try {
      const file = await api.readFile(path);
      setEditorPath(file.path);
      setEditorContent(file.content);
      setEditorDirty(false);
    } catch {
      setEditorPath(path);
      setEditorContent(path.endsWith(".c") ? DEFAULT_C : DEFAULT_V);
      setEditorDirty(false);
    }
  }, []);

  useEffect(() => {
    loadFile(blockFileMap[selectedBlock]);
  }, [selectedBlock, blockFileMap, loadFile]);

  const refreshOutputs = useCallback(async () => {
    const out = await api.pollOutputs();
    setUart(out.uart);
    setLeds(out.leds.leds);
    setServos(out.servos);
    setWaves(out.waves);
    setRefreshKey((k) => k + 1);
  }, []);

  const saveIfDirty = useCallback(async () => {
    if (!editorDirty) return;
    await api.writeFile(editorPath, editorContent);
    setEditorDirty(false);
  }, [editorDirty, editorPath, editorContent]);

  const handleRun = useCallback(async () => {
    try {
      await saveIfDirty();
      await bench.handleRun();
      setHasRun(true);
      await refreshOutputs();
    } catch (e) {
      bench.setError(e as ApiError);
    }
  }, [saveIfDirty, bench, refreshOutputs]);

  const handleStep = useCallback(async () => {
    try {
      await saveIfDirty();
      await bench.handleStep();
      setHasRun(true);
      await refreshOutputs();
    } catch (e) {
      bench.setError(e as ApiError);
    }
  }, [saveIfDirty, bench, refreshOutputs]);

  const handleReset = useCallback(async () => {
    await bench.handleReset();
    setHasRun(false);
    setUart("");
    setLeds(0);
    setServos({ duty: [0, 0, 0, 0] });
    setWaves({ clk: [], gpu_busy: [], led0: [], pwm0: [] });
    setRefreshKey((k) => k + 1);
  }, [bench]);

  const handleBlockSelect = useCallback(
    async (id: BlockId) => {
      if (editorDirty) {
        try {
          await api.writeFile(editorPath, editorContent);
          setEditorDirty(false);
        } catch {
          /* keep editing */
        }
      }
      setSelectedBlock(id);
    },
    [editorDirty, editorPath, editorContent],
  );

  const handleEditorChange = useCallback((value: string) => {
    setEditorContent(value);
    setEditorDirty(true);
  }, []);

  const handleButtonToggle = useCallback(async (down: boolean) => {
    setButtonDown(down);
    try {
      await api.setButton(down);
    } catch {
      /* optional endpoint */
    }
  }, []);

  const handleGateCount = useCallback(async () => {
    setGateOpen(true);
    setGateLoading(true);
    setGateError(undefined);
    setGateCells(undefined);
    setGateWires(undefined);
    try {
      await saveIfDirty();
      const result = await api.gateCount();
      setGateCells(result.cells);
      setGateWires(result.wires);
      if (!result.ok) setGateError(result.error ?? result.message ?? "Gate count failed");
    } catch (e) {
      const err = e as ApiError;
      setGateError(err.message ?? err.error);
    } finally {
      setGateLoading(false);
    }
  }, [saveIfDirty]);

  const handleExportFpga = useCallback(async () => {
    try {
      await saveIfDirty();
      const blob = await api.exportFpga();
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "tiny-gpu-bench.bin";
      a.click();
      URL.revokeObjectURL(url);
    } catch (e) {
      bench.setError(e as ApiError);
    }
  }, [saveIfDirty, bench]);

  const completeOnboarding = useCallback(() => {
    localStorage.setItem(ONBOARDING_KEY, "1");
    setOnboardingOpen(false);
  }, []);

  if (bench.loading) {
    return <BenchSkeleton />;
  }

  return (
    <TooltipProvider delayDuration={300}>
      <div className="flex h-screen w-full flex-col bg-void">
        <TopBar
          cycles={bench.cycles}
          passState={bench.passState}
          running={bench.running}
          onRun={handleRun}
          onStep={handleStep}
          onReset={handleReset}
          onGateCount={handleGateCount}
          onExportFpga={handleExportFpga}
          onOpenMcp={() => bench.setDrawerOpen(true)}
        />

        <div className="flex min-h-0 flex-1">
          <LeftRail selected={selectedBlock} onSelect={handleBlockSelect} />
          <main className="flex min-w-0 flex-1 flex-col p-3">
            <CodeEditor
              path={editorPath}
              value={editorContent}
              onChange={handleEditorChange}
            />
          </main>
          <RightRail
            leds={leds}
            servos={servos}
            refreshKey={refreshKey}
            hasRun={hasRun}
            buttonDown={buttonDown}
            onButtonToggle={handleButtonToggle}
          />
        </div>

        <BottomPanel waves={waves} uart={uart} hasRun={hasRun} />

        <McpDrawer
          open={bench.drawerOpen}
          onOpenChange={bench.setDrawerOpen}
          state={bench.mcpState}
        />

        <OnboardingOverlay open={onboardingOpen} onComplete={completeOnboarding} />

        <GateCountDialog
          open={gateOpen}
          onOpenChange={setGateOpen}
          loading={gateLoading}
          cells={gateCells}
          wires={gateWires}
          error={gateError}
        />

        <ErrorSlideOver error={bench.error} onDismiss={() => bench.setError(null)} />
      </div>
    </TooltipProvider>
  );
}
