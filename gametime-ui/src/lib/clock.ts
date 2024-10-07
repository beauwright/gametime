type Clock = {
    secondsRemain: number;
    increment: number;
};

type ClockState = Map<string, Clock>;
