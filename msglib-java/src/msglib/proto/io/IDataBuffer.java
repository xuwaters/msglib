package msglib.proto.io;

import java.io.IOException;

public interface IDataBuffer {
    int Read(byte[] buffer, int offset, int length) throws IOException;

    void Write(byte[] buffer, int offset, int length) throws IOException;
}
