import { useEffect, useRef } from "react";
import type { QuizPublicState } from "../../types/quizTypes";
import { quizSoundEngine } from "./quizSoundEngine";

/**
 * Big Screen audio cues driven by the quiz state and the locally ticking timer.
 * Attach ONLY in the projector component — never in the host panel or Herald
 * screens (WebSocket latency would cause echo in the hall).
 *
 * All triggers are edge-based against refs, so a page reload (F5) or the first
 * populated render never replays sounds retroactively.
 */
export function useQuizSounds(state: QuizPublicState | null, remainingMs: number) {
    const initializedRef = useRef(false);
    const questionKeyRef = useRef("");
    const lastTickedSecRef = useRef<number | null>(null);
    const timeUpPlayedRef = useRef(false);
    const prevRemainingMsRef = useRef(0);
    const prevCancelledRef = useRef(false);
    const prevBuzzCountRef = useRef(0);

    useEffect(() => {
        if (!state) return;

        const activeQuestion = state.tiebreak_active
            ? state.tiebreak_question
            : state.current_question;
        const questionKey = `${state.tiebreak_active ? "tb" : state.current_round}:${activeQuestion}`;

        // First populated state after mount/reload: record the baseline silently.
        if (!initializedRef.current) {
            initializedRef.current = true;
            questionKeyRef.current = questionKey;
            prevRemainingMsRef.current = remainingMs;
            prevCancelledRef.current = !!state.question?.cancelled;
            prevBuzzCountRef.current = state.buzz_queue.length;
            return;
        }

        // Question change: chime once, reset per-question cue state.
        if (questionKey !== questionKeyRef.current) {
            questionKeyRef.current = questionKey;
            lastTickedSecRef.current = null;
            timeUpPlayedRef.current = false;
            prevBuzzCountRef.current = state.buzz_queue.length;
            if (activeQuestion > 0) {
                quizSoundEngine.playQuestionStart();
            }
        }

        // Buzzer press: queue grew on the same question (round 1 / tie-break).
        if (state.buzz_queue.length > prevBuzzCountRef.current && activeQuestion > 0) {
            quizSoundEngine.playBuzz();
        }
        prevBuzzCountRef.current = state.buzz_queue.length;

        const displaySec = Math.ceil(remainingMs / 1000);

        // Final-10-seconds countdown: strictly decreasing seconds while running,
        // so WS resyncs and pause/resume never double-tick.
        if (
            state.timer.running &&
            activeQuestion > 0 &&
            displaySec >= 1 &&
            displaySec <= 10 &&
            lastTickedSecRef.current !== displaySec &&
            (lastTickedSecRef.current === null || displaySec < lastTickedSecRef.current)
        ) {
            lastTickedSecRef.current = displaySec;
            quizSoundEngine.playCountdownTick(displaySec);
        }

        // Question cancelled: annulment cue on the cancelled flag flip. Runs
        // before the time-up check and suppresses it — cancelling zeroes the
        // timer in the same broadcast.
        const cancelled = !!state.question?.cancelled;
        if (cancelled && !prevCancelledRef.current && activeQuestion > 0) {
            timeUpPlayedRef.current = true;
            quizSoundEngine.playQuestionCancelled();
        }
        prevCancelledRef.current = cancelled;

        // Time up / question closed: exactly once per question on a live
        // > 0 -> 0 transition (natural expiry or the host closing the question,
        // both zero the timer). A reload at 0:00 stays silent via the baseline.
        if (
            activeQuestion > 0 &&
            remainingMs <= 0 &&
            prevRemainingMsRef.current > 0 &&
            !timeUpPlayedRef.current
        ) {
            timeUpPlayedRef.current = true;
            quizSoundEngine.playTimeUp();
        }
        prevRemainingMsRef.current = remainingMs;
    }, [state, remainingMs]);
}
