import { useEffect, useRef, useState } from "react";
import { axiosGetQuizState, connectQuizSocket } from "../../requests/quizRequests";
import type { QuizPublicState } from "../../types/quizTypes";

/**
 * Subscribes to the quiz public state: initial fetch, live WS updates,
 * and a polling fallback every 5 seconds in case the socket drops.
 */
export function useQuizState(eventId: string | undefined) {
    const [state, setState] = useState<QuizPublicState | null>(null);
    const [connected, setConnected] = useState(false);
    const socketRef = useRef<WebSocket | null>(null);

    useEffect(() => {
        if (!eventId) return;

        let cancelled = false;

        axiosGetQuizState(eventId)
            .then((s) => {
                if (!cancelled) setState(s);
            })
            .catch(() => {
                /* keep polling */
            });

        const socket = connectQuizSocket(eventId, (s) => {
            if (!cancelled) {
                setState(s);
                setConnected(true);
            }
        });
        socketRef.current = socket;

        socket.onclose = () => {
            if (!cancelled) setConnected(false);
        };
        socket.onerror = () => {
            socket.close();
        };

        const poll = setInterval(() => {
            axiosGetQuizState(eventId)
                .then((s) => {
                    if (!cancelled) setState(s);
                })
                .catch(() => {
                    /* transient network error */
                });
        }, 5000);

        return () => {
            cancelled = true;
            clearInterval(poll);
            socket.close();
            socketRef.current = null;
        };
    }, [eventId]);

    return { state, connected };
}
