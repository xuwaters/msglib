package msglib.proto.io;

import java.io.IOException;

public interface ITextBuffer {
    String Read() throws IOException;
    void Write(String text) throws IOException;
}
