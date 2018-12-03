package msglib.proto.io;

import java.io.IOException;

public class MException extends IOException {

    private static final long serialVersionUID = 1L;

    public MException() {

    }

    public MException(String message) {
        super(message);
    }

    public MException(String message, Throwable cause) {
        super(message, cause);
    }
}
