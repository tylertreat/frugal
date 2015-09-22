struct Event {
    1: i64 ID,
    2: string Message
}

service Linking {
    void done(1:Event event),
}
