/**
 * Web Audio API sound engine for the Big Screen (projector).
 * Synthesizes broadcast-style cues in the browser with generated studio reverb,
 * layered detuned harmonics, percussive transients, and a tuned master limiter.
 */

const MUTE_STORAGE_KEY = "quiz_sound_muted";
const SILENCE = 0.0001;

interface ToneOptions {
    freq: number;
    freqEnd?: number;
    freqRampTime?: number;
    type: OscillatorType;
    gain: number;
    attack: number;
    hold?: number; // sustain duration at peak gain before decay
    decay: number;
    start: number;
    detune?: number; // cents
    filterStart?: number; // Hz lowpass start
    filterEnd?: number; // Hz lowpass end (envelope sweep)
    filterQ?: number;
    reverb?: number; // 0..1 wet send amount
}

interface NoiseOptions {
    gain: number;
    decay: number;
    start: number;
    bandpass: number; // Hz
    q?: number;
    reverb?: number;
}

class QuizSoundEngine {
    private ctx: AudioContext | null = null;
    private dryBus: GainNode | null = null;
    private reverbConvolver: ConvolverNode | null = null;
    private masterGain: GainNode | null = null;
    private noiseBuffer: AudioBuffer | null = null;
    private muted = false;
    private volume = 1.0;

    constructor() {
        try {
            this.muted = localStorage.getItem(MUTE_STORAGE_KEY) === "1";
        } catch {
            /* storage unavailable */
        }
    }

    /** Must be called from a user gesture; resumes the context. */
    async unlock(): Promise<boolean> {
        if (!this.ctx) {
            const Ctor =
                window.AudioContext ??
                (window as unknown as { webkitAudioContext?: typeof AudioContext })
                    .webkitAudioContext;
            if (!Ctor) return false;

            const ctx = new Ctor();
            this.ctx = ctx;

            // Gentle master limiter instead of default crushing compressor
            const limiter = ctx.createDynamicsCompressor();
            limiter.threshold.value = -8;
            limiter.knee.value = 10;
            limiter.ratio.value = 4;
            limiter.attack.value = 0.003;
            limiter.release.value = 0.16;

            this.masterGain = ctx.createGain();
            this.masterGain.gain.value = this.muted ? 0 : this.volume;

            this.dryBus = ctx.createGain();
            this.dryBus.gain.value = 1.0;

            // Synthetic warm studio hall reverb
            this.reverbConvolver = ctx.createConvolver();
            this.reverbConvolver.buffer = this.createReverbImpulse(ctx, 1.15, 2.8);

            const reverbReturn = ctx.createGain();
            reverbReturn.gain.value = 0.45;

            this.dryBus.connect(this.masterGain);
            this.reverbConvolver.connect(reverbReturn);
            reverbReturn.connect(this.masterGain);

            this.masterGain.connect(limiter);
            limiter.connect(ctx.destination);

            this.noiseBuffer = this.createNoiseBuffer(ctx, 0.4);
        }

        if (this.ctx.state === "suspended") {
            try {
                await this.ctx.resume();
            } catch {
                /* still blocked */
            }
        }
        return this.isUnlocked();
    }

    isUnlocked(): boolean {
        return !!this.ctx && this.ctx.state === "running";
    }

    isMuted(): boolean {
        return this.muted;
    }

    setMuted(muted: boolean) {
        this.muted = muted;
        try {
            localStorage.setItem(MUTE_STORAGE_KEY, muted ? "1" : "0");
        } catch {
            /* storage unavailable */
        }
        if (this.ctx && this.masterGain) {
            this.masterGain.gain.setTargetAtTime(
                muted ? 0 : this.volume,
                this.ctx.currentTime,
                0.02,
            );
        }
    }

    /** Question start: bright 4-note broadcast chime with bell overtones & mallet strike. */
    playQuestionStart() {
        const ctx = this.ready();
        if (!ctx) return;
        const t = ctx.currentTime;

        const notes = [
            { freq: 523.25, delay: 0.0, gain: 0.42 }, // C5
            { freq: 659.25, delay: 0.055, gain: 0.44 }, // E5
            { freq: 783.99, delay: 0.11, gain: 0.48 }, // G5
            { freq: 1046.5, delay: 0.165, gain: 0.52 }, // C6
        ];

        // Warm foundation chord swell underneath
        this.tone({
            freq: 261.63, // C4
            type: "triangle",
            gain: 0.32,
            attack: 0.02,
            hold: 0.15,
            decay: 1.25,
            start: t,
            filterStart: 1600,
            filterEnd: 600,
            reverb: 0.45,
        });

        for (const n of notes) {
            const start = t + n.delay;
            // Rich detuned bell body
            this.tone({
                freq: n.freq,
                type: "triangle",
                gain: n.gain,
                attack: 0.006,
                hold: 0.04,
                decay: 1.05,
                start,
                detune: -5,
                filterStart: 4200,
                filterEnd: 1400,
                reverb: 0.42,
            });
            this.tone({
                freq: n.freq,
                type: "sine",
                gain: n.gain * 0.85,
                attack: 0.006,
                hold: 0.05,
                decay: 1.15,
                start,
                detune: 5,
                reverb: 0.42,
            });
            // Metallic vibraphone/bell partial (2.76x harmonic)
            this.tone({
                freq: n.freq * 2.76,
                type: "sine",
                gain: n.gain * 0.16,
                attack: 0.003,
                decay: 0.28,
                start,
                reverb: 0.35,
            });
        }

        // Crisp mallet click on entry and peak note
        this.noiseBurst({ gain: 0.18, decay: 0.022, start: t, bandpass: 2800, q: 2.0, reverb: 0.2 });
        this.noiseBurst({ gain: 0.22, decay: 0.025, start: t + 0.165, bandpass: 3400, q: 2.2, reverb: 0.25 });
    }

    /** Final 10s countdown: studio clock click + sub-heartbeat; urgent bell ping on 3..1. */
    playCountdownTick(secondsLeft: number) {
        const ctx = this.ready();
        if (!ctx) return;
        const t = ctx.currentTime;
        const urgent = secondsLeft <= 3;

        // 1. Crisp wooden clock tick (noise + fast pitch drop)
        this.noiseBurst({
            gain: urgent ? 0.35 : 0.24,
            decay: 0.02,
            start: t,
            bandpass: urgent ? 2800 : 2100,
            q: 2.8,
            reverb: 0.1,
        });
        this.tone({
            freq: urgent ? 1500 : 1150,
            freqEnd: urgent ? 680 : 480,
            freqRampTime: 0.028,
            type: "triangle",
            gain: urgent ? 0.55 : 0.42,
            attack: 0.002,
            decay: 0.06,
            start: t,
            reverb: 0.08,
        });

        // 2. Tense studio heartbeat sub-thump (cuts through hall acoustics)
        this.tone({
            freq: urgent ? 115 : 95,
            freqEnd: 48,
            freqRampTime: 0.09,
            type: "sine",
            gain: urgent ? 0.72 : 0.55,
            attack: 0.004,
            decay: 0.13,
            start: t,
            reverb: 0.06,
        });

        // 3. Urgent high harmonic ping for 3, 2, 1
        if (urgent) {
            const pingFreq = secondsLeft === 1 ? 1567.98 : 1318.51; // G6 on 1, E6 on 3 & 2
            this.tone({
                freq: pingFreq,
                type: "sine",
                gain: 0.38,
                attack: 0.004,
                hold: 0.025,
                decay: 0.22,
                start: t,
                reverb: 0.28,
            });
        }
    }

    /** Time up: rhythmic 4-note broadcast cadence ("Tu-du — tu-duuu!") with deep finale. */
    playTimeUp() {
        const ctx = this.ready();
        if (!ctx) return;
        const t = ctx.currentTime;

        const notes = [
            // Первая пара: «Ту-ду»
            { freq: 392.0, start: t + 0.0, hold: 0.055, decay: 0.11, reverb: 0.18, final: false }, // G4
            { freq: 329.63, start: t + 0.13, hold: 0.075, decay: 0.17, reverb: 0.22, final: false }, // E4
            // Вторая пара: «ту-дууу!»
            { freq: 293.66, start: t + 0.34, hold: 0.055, decay: 0.11, reverb: 0.18, final: false }, // D4
            { freq: 261.63, start: t + 0.47, hold: 0.32, decay: 0.95, reverb: 0.42, final: true }, // C4
        ];

        for (const n of notes) {
            // Плотный духовой/синт-слой (пара расстроенных пил через фильтр)
            for (const detune of [-6, 6]) {
                this.tone({
                    freq: n.freq,
                    type: "sawtooth",
                    gain: n.final ? 0.28 : 0.24,
                    attack: 0.008,
                    hold: n.hold,
                    decay: n.decay,
                    start: n.start,
                    detune,
                    filterStart: n.final ? 2800 : 2400,
                    filterEnd: n.final ? 700 : 1100,
                    filterQ: 1.4,
                    reverb: n.reverb,
                });
            }

            // Тело ноты (октава ниже)
            this.tone({
                freq: n.freq / 2,
                type: "triangle",
                gain: n.final ? 0.6 : 0.46,
                attack: 0.008,
                hold: n.hold,
                decay: n.decay * 0.95,
                start: n.start,
                filterStart: 1000,
                filterEnd: 320,
                reverb: n.reverb * 0.7,
            });

            // Чёткий перкуссионный щелчок атаки, чтобы ритм читался в шумном зале
            this.noiseBurst({
                gain: n.final ? 0.24 : 0.16,
                decay: 0.03,
                start: n.start,
                bandpass: 1400,
                q: 1.4,
                reverb: 0.2,
            });

            // Финальный акцент: глубокий саб + лёгкий квинтовый обертон
            if (n.final) {
                this.tone({
                    freq: n.freq / 2,
                    type: "sine",
                    gain: 0.52,
                    attack: 0.01,
                    hold: n.hold,
                    decay: n.decay,
                    start: n.start,
                    reverb: 0.15,
                });
                this.tone({
                    freq: 392.0, // G4 квинта к финальной C4 для «студийного» объёма
                    type: "triangle",
                    gain: 0.18,
                    attack: 0.015,
                    hold: 0.22,
                    decay: 0.85,
                    start: n.start,
                    filterStart: 2000,
                    filterEnd: 600,
                    reverb: 0.45,
                });
            }
        }
    }

    /** Question cancelled: dark descending "annulled" glide, clearly distinct from the time-up cadence. */
    playQuestionCancelled() {
        const ctx = this.ready();
        if (!ctx) return;
        const t = ctx.currentTime;

        // Dampened thud marking the annulment
        this.noiseBurst({ gain: 0.3, decay: 0.05, start: t, bandpass: 900, q: 1.2, reverb: 0.2 });

        // Descending detuned saw glide (Eb4 -> Bb3), muted and dark
        for (const detune of [-7, 7]) {
            this.tone({
                freq: 311.13,
                freqEnd: 233.08,
                freqRampTime: 0.35,
                type: "sawtooth",
                gain: 0.22,
                attack: 0.012,
                hold: 0.1,
                decay: 0.65,
                start: t,
                detune,
                filterStart: 1400,
                filterEnd: 320,
                filterQ: 1.2,
                reverb: 0.3,
            });
        }

        // Dissonant minor-second tail (A3 + Bb3) — "the question is annulled"
        for (const f of [220, 233.08]) {
            this.tone({
                freq: f,
                type: "triangle",
                gain: 0.26,
                attack: 0.02,
                hold: 0.18,
                decay: 0.7,
                start: t + 0.3,
                filterStart: 900,
                filterEnd: 260,
                reverb: 0.35,
            });
        }

        // Low sub drop underneath
        this.tone({
            freq: 155.56, // Eb3
            freqEnd: 98, // G2
            freqRampTime: 0.4,
            type: "sine",
            gain: 0.4,
            attack: 0.015,
            hold: 0.08,
            decay: 0.6,
            start: t,
            reverb: 0.12,
        });
    }

    /** ESP32 buzzer press: punchy game-show lock-in chord with crisp attack. */
    playBuzz() {
        const ctx = this.ready();
        if (!ctx) return;
        const t = ctx.currentTime;

        // Percussive lock-in click
        this.noiseBurst({ gain: 0.28, decay: 0.02, start: t, bandpass: 2400, q: 1.8, reverb: 0.15 });

        // Bright power-fifth synth hit (F#5 + C#6) with slight upward snap
        const freqs = [739.99, 1108.73];
        for (const f of freqs) {
            this.tone({
                freq: f * 0.92,
                freqEnd: f,
                freqRampTime: 0.03,
                type: "sawtooth",
                gain: 0.28,
                attack: 0.004,
                hold: 0.055,
                decay: 0.24,
                start: t,
                filterStart: 3800,
                filterEnd: 1500,
                reverb: 0.22,
            });
            this.tone({
                freq: f,
                type: "triangle",
                gain: 0.35,
                attack: 0.004,
                hold: 0.06,
                decay: 0.26,
                start: t,
                detune: 6,
                reverb: 0.22,
            });
        }
    }

    private ready(): AudioContext | null {
        return this.ctx && this.ctx.state === "running" ? this.ctx : null;
    }

    private tone(o: ToneOptions) {
        if (!this.ctx || !this.dryBus) return;
        const ctx = this.ctx;
        const osc = ctx.createOscillator();
        const env = ctx.createGain();

        osc.type = o.type;
        osc.frequency.setValueAtTime(o.freq, o.start);
        if (o.detune !== undefined) {
            osc.detune.setValueAtTime(o.detune, o.start);
        }
        if (o.freqEnd !== undefined) {
            osc.frequency.exponentialRampToValueAtTime(
                Math.max(o.freqEnd, 20),
                o.start + (o.freqRampTime ?? 0.07),
            );
        }

        const peakTime = o.start + o.attack;
        const holdEnd = peakTime + (o.hold ?? 0);
        const endTime = holdEnd + o.decay;

        env.gain.setValueAtTime(SILENCE, o.start);
        env.gain.linearRampToValueAtTime(o.gain, peakTime);
        if (o.hold) {
            env.gain.setValueAtTime(o.gain, holdEnd);
        }
        env.gain.exponentialRampToValueAtTime(SILENCE, endTime);

        let node: AudioNode = osc;
        if (o.filterStart !== undefined) {
            const filter = ctx.createBiquadFilter();
            filter.type = "lowpass";
            filter.Q.value = o.filterQ ?? 0.8;
            filter.frequency.setValueAtTime(o.filterStart, o.start);
            if (o.filterEnd !== undefined) {
                filter.frequency.exponentialRampToValueAtTime(
                    Math.max(o.filterEnd, 60),
                    endTime,
                );
            }
            node.connect(filter);
            node = filter;
        }

        node.connect(env);
        env.connect(this.dryBus);

        if (o.reverb && o.reverb > 0 && this.reverbConvolver) {
            const wet = ctx.createGain();
            wet.gain.value = o.reverb;
            env.connect(wet);
            wet.connect(this.reverbConvolver);
        }

        osc.start(o.start);
        osc.stop(endTime + 0.05);
    }

    private noiseBurst(o: NoiseOptions) {
        if (!this.ctx || !this.dryBus || !this.noiseBuffer) return;
        const ctx = this.ctx;
        const src = ctx.createBufferSource();
        src.buffer = this.noiseBuffer;

        const filter = ctx.createBiquadFilter();
        filter.type = "bandpass";
        filter.frequency.value = o.bandpass;
        filter.Q.value = o.q ?? 1.5;

        const env = ctx.createGain();
        env.gain.setValueAtTime(o.gain, o.start);
        env.gain.exponentialRampToValueAtTime(SILENCE, o.start + o.decay);

        src.connect(filter);
        filter.connect(env);
        env.connect(this.dryBus);

        if (o.reverb && o.reverb > 0 && this.reverbConvolver) {
            const wet = ctx.createGain();
            wet.gain.value = o.reverb;
            env.connect(wet);
            wet.connect(this.reverbConvolver);
        }

        src.start(o.start);
        src.stop(o.start + o.decay + 0.03);
    }

    /** Generates a stereo decaying noise buffer for studio hall reverb. */
    private createReverbImpulse(ctx: AudioContext, duration: number, decayPow: number): AudioBuffer {
        const length = Math.floor(ctx.sampleRate * duration);
        const impulse = ctx.createBuffer(2, length, ctx.sampleRate);
        for (let ch = 0; ch < 2; ch++) {
            const data = impulse.getChannelData(ch);
            let lp = 0;
            for (let i = 0; i < length; i++) {
                const env = Math.pow(1 - i / length, decayPow);
                const white = Math.random() * 2 - 1;
                // Simple 1-pole lowpass so the tail sounds warm, not metallic
                lp = lp * 0.72 + white * 0.28;
                data[i] = lp * env;
            }
        }
        return impulse;
    }

    private createNoiseBuffer(ctx: AudioContext, duration: number): AudioBuffer {
        const length = Math.floor(ctx.sampleRate * duration);
        const buf = ctx.createBuffer(1, length, ctx.sampleRate);
        const data = buf.getChannelData(0);
        for (let i = 0; i < length; i++) {
            data[i] = Math.random() * 2 - 1;
        }
        return buf;
    }
}

export const quizSoundEngine = new QuizSoundEngine();