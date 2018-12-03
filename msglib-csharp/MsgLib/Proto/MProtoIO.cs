using System;
using System.Collections.Generic;
using System.Text;

namespace MsgLib
{
    internal interface IDataBuffer
    {
        int Read(byte[] buffer, int offset, int length);
        void Write(byte[] buffer, int offset, int length);
    }

    internal abstract class MProto
    {
        protected IDataBuffer databuffer;
        protected MProto(IDataBuffer reader)
        {
            this.databuffer = reader;
        }
        internal IDataBuffer DataBuffer { get { return this.databuffer; } }

        internal abstract MAction ReadActionBegin();
        internal abstract MStruct ReadStructBegin();
        internal abstract MField ReadFieldBegin();
        internal abstract MMap ReadMapBegin();
        internal abstract MList ReadListBegin();
        internal abstract MSet ReadSetBegin();
        internal abstract bool ReadBool();
        internal abstract sbyte ReadByte();
        internal abstract short ReadI16();
        internal abstract int ReadI32();
        internal abstract long ReadI64();
        internal abstract float ReadFloat();
        internal abstract double ReadDouble();
        internal abstract byte[] ReadBinary();

        internal virtual string ReadString()
        {
            var buf = ReadBinary();
            return Encoding.UTF8.GetString(buf, 0, buf.Length);
        }

        internal void ReadAll(byte[] buffer, int offset, int length)
        {
            int len = databuffer.Read(buffer, offset, length);
            if (len < length)
                throw new MException("Read EOF");
        }

        internal abstract void WriteActionBegin(MAction action);
        internal abstract void WriteStructBegin(MStruct struc);
        internal abstract void WriteFieldBegin(MField field);

        internal virtual void WriteFieldStop()
        {
            WriteFieldBegin(new MField("", MType.Null, 0));
        }

        internal abstract void WriteMapBegin(MMap map);
        internal abstract void WriteListBegin(MList list);
        internal abstract void WriteSetBegin(MSet set);
        internal abstract void WriteBool(bool b);
        internal abstract void WriteByte(sbyte b);
        internal abstract void WriteI16(short i16);
        internal abstract void WriteI32(int i32);
        internal abstract void WriteI64(long i64);
        internal abstract void WriteFloat(float f);
        internal abstract void WriteDouble(double d);
        internal abstract void WriteBinary(byte[] binary, int offset, int length);
        
        internal virtual void WriteBinary(byte[] binary)
        {
            WriteBinary(binary, 0, binary.Length);
        }

        internal virtual void WriteString(string str)
        {
            WriteBinary(Encoding.UTF8.GetBytes(str));
        }
    }

}
