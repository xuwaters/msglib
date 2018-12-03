package msglib.proto.io;

import java.io.IOException;

import msglib.stream.MemoryStream;

public class MemoryDataBuffer extends MemoryStream implements IDataBuffer {

    public MemoryDataBuffer() {
    }

    public MemoryDataBuffer(byte[] data, boolean writable) {
        super(data, writable);
    }

    public MemoryDataBuffer(byte[] data, int offset, int length, boolean writable) {
        super(data, offset, length, writable, false);
    }

    @Override
    public int Read(byte[] buffer, int offset, int length) throws IOException {
        return super.read(buffer, offset, length);
    }

    @Override
    public void Write(byte[] buffer, int offset, int length) throws IOException {
        super.write(buffer, offset, length);
    }

}
