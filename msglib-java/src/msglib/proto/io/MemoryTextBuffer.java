package msglib.proto.io;

import java.io.IOException;
import java.util.ArrayList;

/**
 * Do not support null string
 */
public class MemoryTextBuffer implements ITextBuffer {

    private final ArrayList<String> buffer;
    private int readIndex = 0;

    public MemoryTextBuffer() {
        this.buffer = new ArrayList<String>();
    }

    public MemoryTextBuffer(ArrayList<String> buff) {
        if (buff != null) {
            throw new IllegalArgumentException("null buffer");
        }
        this.buffer = buff;
    }

    public MemoryTextBuffer(String data) {
        this.buffer = parseText(data);
    }

    @Override
    public String Read() throws IOException {
        if (readIndex >= 0 && readIndex < buffer.size()) {
            String str = buffer.get(readIndex);
            readIndex++;
            if (str == null) {
                str = "";
            }
            return str;
        }
        return null;
    }

    @Override
    public void Write(String text) throws IOException {
        if (text == null) {
            text = "";
        }
        buffer.add(text);
    }

    public String getText() {
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < this.buffer.size(); i++) {
            String val = this.buffer.get(i);
            if (val == null) {
                val = "";
            }
            else {
                val = val.replace("\\", "\\\\");
                val = val.replace(";", "\\;");
            }
            if (i > 0) {
                sb.append(';');
            }
            sb.append(val);
        }
        return sb.toString();
    }

    private static ArrayList<String> parseText(String text) {
        ArrayList<String> list = new ArrayList<String>();
        int start = 0;
        int end = 0;
        while (start < text.length()) {
            StringBuilder curr = new StringBuilder();
            for (end = start; end < text.length(); ) {
                char c = text.charAt(end);
                if (c == ';') {
                    list.add(curr.toString());
                    ++end;
                    break;
                }
                //
                if (c == '\\') {
                    if (end + 1 >= text.length()) {
                        curr.append(c);
                        ++end;
                        break;
                    }
                    // next
                    char n = text.charAt(end + 1);
                    if (n == '\\' || n == ';') {
                        ++end;
                        curr.append(n);
                    }
                    else {
                        curr.append(c);
                    }
                }
                else {
                    curr.append(c);
                }
                
                // 
                ++end;
                if (end >= text.length()) {
                    list.add(curr.toString());
                    break;
                }
            }
            //
            start = end;
        }
        return list;
    }
}
