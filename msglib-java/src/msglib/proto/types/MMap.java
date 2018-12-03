package msglib.proto.types;

public class MMap {

    public int KeyType;
    public int ValueType;
    public int Count;

    public MMap() {
    }

    public MMap(int keytype, int valuetype, int count) {
        this.KeyType = keytype;
        this.ValueType = valuetype;
        this.Count = count;
    }

}
