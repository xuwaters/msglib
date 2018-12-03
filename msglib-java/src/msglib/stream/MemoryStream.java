package msglib.stream;

import java.io.IOException;
import java.util.Arrays;

public class MemoryStream extends Stream {

    private final int MAX_LENGTH = 0x7FFFFFFF;

    private byte[] buffer;

    private int capacity;

    private int length;

    private int origin;

    private int position;

    private boolean opened;

    private boolean exposable;

    private boolean expandable;

    private boolean writable;

    private final Object kMutex = new Object();

    public MemoryStream() {
        this(0);
    }

    public MemoryStream(int capacity) {
        if (capacity < 0) {
            throw new IllegalArgumentException("capacity < 0");
        }
        this.buffer = new byte[capacity];
        this.capacity = capacity;
        this.expandable = true;
        this.writable = true;
        this.exposable = true;
        this.origin = 0;
        this.opened = true;
    }

    public MemoryStream(byte[] buffer, boolean writable) {
        if (buffer == null) {
            throw new IllegalArgumentException("buffer == null");
        }
        this.buffer = buffer;
        this.length = this.capacity = buffer.length;
        this.writable = writable;
        this.exposable = false;
        this.origin = 0;
        this.opened = true;
    }

    public MemoryStream(byte[] buffer, int offset, int length, boolean writable, boolean publiclyVisible) {
        checkBuffer(buffer, offset, length);
        this.buffer = buffer;
        this.origin = this.position = offset;
        this.length = this.capacity = offset + length;
        this.writable = writable;
        this.exposable = publiclyVisible;
        this.expandable = false;
        this.opened = true;
    }

    @Override
    public boolean canRead() throws IOException {
        return this.opened;
    }

    @Override
    public boolean canSeek() throws IOException {
        return true;
    }

    @Override
    public boolean canWrite() throws IOException {
        return this.writable;
    }

    @Override
    public void close() throws IOException {
        this.opened = false;
        this.writable = false;
        this.expandable = false;
    }

    @Override
    public void flush() throws IOException {
        // nothing
    }

    private boolean ensureCapacity(int value) throws IOException {
        if (value < 0) {
            throw new IOException("invalid capacity value");
        }
        if (value <= this.capacity) {
            return false;
        }
        if (value < 0x100) {
            value = 0x100;
        }
        if (value < (this.capacity << 1)) {
            value = (this.capacity << 1);
        }
        setCapacity(value);
        return true;
    }

    public byte[] getBuffer() {
        if (!this.exposable) {
            throw new IllegalStateException("buffer not exposable");
        }
        return this.buffer;
    }

    public byte[] toArray() {
        synchronized (kMutex) {
            return Arrays.copyOfRange(this.buffer, this.origin, this.length);
        }
    }

    private void checkOpen() throws IOException {
        if (!this.opened) {
            throw new IOException("stream closed");
        }
    }

    private void checkBuffer(byte[] buffer, int offset, int length) {
        if (buffer == null) {
            throw new IllegalArgumentException("buffer == null");
        }
        if (offset < 0) {
            throw new IllegalArgumentException("index < 0");
        }
        if (length < 0) {
            throw new IllegalArgumentException("count < 0");
        }
        if ((buffer.length - offset) < length) {
            throw new IllegalArgumentException("buffer.length < index + count");
        }
    }

    @Override
    public int read(byte[] buffer, int offset, int length) throws IOException {
        synchronized (kMutex) {
            checkOpen();
            checkBuffer(buffer, offset, length);
            int num = this.length - this.position;
            if (num > length) {
                num = length;
            }
            if (num <= 0) {
                return 0;
            }
            System.arraycopy(this.buffer, this.position, buffer, offset, num);
            this.position += num;
            return num;
        }
    }

    @Override
    public int read() throws IOException {
        synchronized (kMutex) {
            checkOpen();
            if (this.position >= this.length) {
                return -1;
            }
            int value = (this.buffer[this.position++] & 0xFF);
            return value;
        }
    }

    @Override
    public long seek(long offset, SeekOrigin origin) throws IOException {
        synchronized (kMutex) {
            checkOpen();
            if (offset > MAX_LENGTH) {
                throw new IllegalArgumentException("offset too large");
            }
            switch (origin) {
            case Begin:
                if (offset < 0) {
                    throw new IllegalArgumentException("seek before begin");
                }
                this.position = this.origin + (int) offset;
                break;
            case Current:
                if ((this.position + (int) offset) < this.origin) {
                    throw new IllegalArgumentException("seek before begin");
                }
                this.position += (int) offset;
                break;
            case End:
                if ((this.length + (int) offset) < this.origin) {
                    throw new IllegalArgumentException("seek before begin");
                }
                this.position = this.length + (int) offset;
                break;
            default:
                throw new IllegalArgumentException("invalid SeekOrigin");
            }
            return this.position;
        }
    }

    private void checkWritable() {
        if (!this.writable) {
            throw new IllegalArgumentException("not writable");
        }
    }

    @Override
    public void write(byte[] buffer, int offset, int length) throws IOException {
        synchronized (kMutex) {
            checkOpen();
            checkWritable();
            checkBuffer(buffer, offset, length);
            int num = this.position + length;
            if (num > this.length) {
                boolean flag = (this.position > this.length);
                if (num > this.capacity && ensureCapacity(num)) {
                    flag = false;
                }
                if (flag) {
                    Arrays.fill(this.buffer, this.length, this.position, (byte) 0);
                }
                this.length = num;
            }
            System.arraycopy(buffer, offset, this.buffer, this.position, length);
            this.position = num;
        }
    }

    @Override
    public void write(byte value) throws IOException {
        synchronized (kMutex) {
            checkOpen();
            checkWritable();
            int num = this.position + 1;
            if (num > this.length) {
                boolean flag = (this.position > this.length);
                if (num > this.capacity && ensureCapacity(num)) {
                    flag = false;
                }
                if (flag) {
                    Arrays.fill(this.buffer, this.length, this.position, (byte) 0);
                }
                this.length = num;
            }
            this.buffer[this.position++] = value;
        }
    }

    @Override
    public long getLength() throws IOException {
        synchronized (kMutex) {
            checkOpen();
            return this.length - this.origin;
        }
    }

    @Override
    public void setLength(long value) throws IOException {
        synchronized (kMutex) {
            checkOpen();
            checkWritable();
            if (value > MAX_LENGTH) {
                throw new IllegalArgumentException("length too large");
            }
            if (value < 0 || value > (MAX_LENGTH - this.origin)) {
                throw new IllegalArgumentException("length too large");
            }
            int num = this.origin + (int) value;
            if (!this.ensureCapacity(num) && num > this.length) {
                Arrays.fill(this.buffer, this.length, num, (byte) 0);
            }
            this.length = num;
            if (this.position > num) {
                this.position = num;
            }
        }
    }

    @Override
    public long getPosition() throws IOException {
        synchronized (kMutex) {
            checkOpen();
            return this.position - this.origin;
        }
    }

    @Override
    public void setPosition(long position) throws IOException {
        synchronized (kMutex) {
            checkOpen();
            if (position < 0 || position > MAX_LENGTH) {
                throw new IllegalArgumentException("invalid position");
            }
            this.position = this.origin + (int) position;
        }
    }

    public int getCapacity() throws IOException {
        synchronized (kMutex) {
            checkOpen();
            return this.capacity - this.origin;
        }
    }

    public void setCapacity(int value) throws IOException {
        synchronized (kMutex) {
            checkOpen();
            if (value != this.capacity) {
                if (!this.expandable) {
                    throw new IllegalStateException("buffer not expandable");
                }
                if (value < this.length) {
                    throw new IllegalArgumentException("invalid capacity");
                }
                if (value > 0) {
                    byte[] dst = new byte[value];
                    if (this.length > 0) {
                        System.arraycopy(this.buffer, 0, dst, 0, this.length);
                    }
                    this.buffer = dst;
                }
                else {
                    // never reach here
                    this.buffer = null;
                }
                this.capacity = value;
            }
        }
    }
}
