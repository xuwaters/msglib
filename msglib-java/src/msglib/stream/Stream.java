package msglib.stream;

import java.io.Closeable;
import java.io.IOException;

public abstract class Stream implements Closeable {

    @Override
    public abstract void close() throws IOException;

    public abstract boolean canRead() throws IOException;

    public abstract boolean canSeek() throws IOException;

    public abstract boolean canWrite() throws IOException;

    public abstract long getLength() throws IOException;

    public abstract void setLength(long value) throws IOException;

    public abstract long getPosition() throws IOException;

    public abstract void setPosition(long position) throws IOException;

    public abstract void flush() throws IOException;

    public abstract int read(byte[] buffer, int offset, int length) throws IOException;

    public int read() throws IOException {
        byte[] buffer = new byte[1];
        if (0 == read(buffer, 0, 1)) {
            return -1;
        }
        return buffer[0];
    }

    public abstract long seek(long offset, SeekOrigin origin) throws IOException;

    public abstract void write(byte[] buffer, int offset, int length) throws IOException;

    public void write(byte value) throws IOException {
        byte[] buffer = new byte[] { value };
        write(buffer, 0, 1);
    }
}
