using System;
using System.Collections.Generic;

namespace MsgLib
{
    // thread unsafe
    internal class MBinaryProto : MProto
    {
        internal MBinaryProto(IDataBuffer buffer)
            : base(buffer)
        {
        }

        internal override MAction ReadActionBegin()
        {
            MAction action = new MAction();
            action.ModuleID = (int)ReadVarint32();
            uint actionAndType = ReadVarint32();
            action.ActionID = (int)(actionAndType >> 2);
            action.Type = (MActionType)(actionAndType & 0x03);
            action.SeqID = (int)ReadVarint32();
            return action;
        }

        internal override void WriteActionBegin(MAction action)
        {
            WriteVarint32((uint)action.ModuleID);
            uint actid = (uint)action.ActionID;
            uint type = (uint)action.Type;
            uint actionAndType = ((actid << 2) | (type & 0x03));
            WriteVarint32(actionAndType);
            WriteVarint32((uint)action.SeqID);
        }

        internal override MStruct ReadStructBegin()
        {
            MStruct str = new MStruct();
            return str;
        }

        internal override void WriteStructBegin(MStruct struc)
        {
        }

        internal override MField ReadFieldBegin()
        {
            MField field = new MField();
            uint idAndType = ReadVarint32();
            field.Type = (MType)(idAndType & 0x0F);
            field.ID = (int)(idAndType >> 4);
            return field;
        }

        internal override void WriteFieldBegin(MField field)
        {
            uint id = (uint)field.ID;
            uint type = (uint)field.Type;
            uint idAndType = ((id << 4) | (type & 0x0F));
            WriteVarint32(idAndType);
        }

        internal override MMap ReadMapBegin()
        {
            MMap map = new MMap();
            map.Count = (int)ReadVarint32();
            byte type = (byte)ReadByte();
            map.KeyType = (MType)(type & 0x0F);
            map.ValueType = (MType)(type >> 4);
            return map;
        }

        internal override void WriteMapBegin(MMap map)
        {
            WriteVarint32((uint)map.Count);
            byte val = (byte)map.ValueType;
            byte key = (byte)map.KeyType;
            byte type = (byte)(((val & 0x0F) << 4) | (key & 0x0F));
            WriteByte((sbyte)type);
        }

        internal override MList ReadListBegin()
        {
            MList list = new MList();
            uint countAndType = ReadVarint32();
            list.ElementType = (MType)(countAndType & 0x0F);
            list.Count = (int)(countAndType >> 4);
            return list;
        }

        internal override void WriteListBegin(MList list)
        {
            uint count = (uint)list.Count;
            uint type = (uint)list.ElementType;
            uint countAndType = ((count << 4) | (type & 0x0F));
            WriteVarint32(countAndType);
        }

        internal override MSet ReadSetBegin()
        {
            MList list = ReadListBegin();
            MSet set = new MSet(list.ElementType, list.Count);
            return set;
        }

        internal override void WriteSetBegin(MSet set)
        {
            WriteListBegin(new MList(set.ElementType, set.Count));
        }

        internal override bool ReadBool()
        {
            sbyte val = ReadByte();
            return val == 1;
        }

        internal override void WriteBool(bool b)
        {
            sbyte val = (sbyte)(b ? 1 : 0);
            WriteByte(val);
        }

        byte[] byteRawBuf = new byte[1];
        internal override sbyte ReadByte()
        {
            ReadAll(byteRawBuf, 0, 1);
            return (sbyte)byteRawBuf[0];
        }

        internal override void WriteByte(sbyte b)
        {
            byteRawBuf[0] = (byte)b;
            databuffer.Write(byteRawBuf, 0, 1);
        }

        internal override short ReadI16()
        {
            return (short)zigzagToInt(ReadVarint32());
        }

        internal override void WriteI16(short i16)
        {
            WriteVarint32(intToZigZag(i16));
        }

        internal override int ReadI32()
        {
            return zigzagToInt(ReadVarint32());
        }
        internal override void WriteI32(int i32)
        {
            WriteVarint32(intToZigZag(i32));
        }

        internal override long ReadI64()
        {
            return zigzagToLong(ReadVarint64());
        }

        internal override void WriteI64(long i64)
        {
            WriteVarint64(longToZigzag(i64));
        }

        byte[] floatBuffer = new byte[4];
        internal override float ReadFloat()
        {
            ReadAll(floatBuffer, 0, floatBuffer.Length);
            int val = bytesToInt(floatBuffer);
            float res = BitConverter.ToSingle(BitConverter.GetBytes(val), 0);
            return res;
        }

        internal override void WriteFloat(float f)
        {
            int val = BitConverter.ToInt32(BitConverter.GetBytes(f), 0);
            intToBytes(val, floatBuffer);
            databuffer.Write(floatBuffer, 0, floatBuffer.Length);
        }

        byte[] doubleBuffer = new byte[8];
        internal override double ReadDouble()
        {
            ReadAll(doubleBuffer, 0, 8);
            double res = BitConverter.Int64BitsToDouble(bytesToLong(doubleBuffer));
            return res;
        }

        internal override void WriteDouble(double d)
        {
            longToBytes(BitConverter.DoubleToInt64Bits(d), doubleBuffer);
            databuffer.Write(doubleBuffer, 0, doubleBuffer.Length);
        }

        internal override byte[] ReadBinary()
        {
            int len = (int)ReadVarint32();
            if (len == 0)
                return new byte[0];
            byte[] res = new byte[len];
            ReadAll(res, 0, len);
            return res;
        }

        internal override void WriteBinary(byte[] binary, int offset, int length)
        {
            WriteVarint32((uint)length);
            databuffer.Write(binary, offset, length);
        }

        private void intToBytes(int n, byte[] buf)
        {
            buf[0] = (byte)(n & 0xff);
            buf[1] = (byte)((n >> 8) & 0xff);
            buf[2] = (byte)((n >> 16) & 0xff);
            buf[3] = (byte)((n >> 24) & 0xff);
        }

        private int bytesToInt(byte[] bytes)
        {
            return
              ((bytes[3] & 0xff) << 24) |
              ((bytes[2] & 0xff) << 16) |
              ((bytes[1] & 0xff) << 8) |
              ((bytes[0] & 0xff));
        }

        private void longToBytes(long n, byte[] buf)
        {
            buf[0] = (byte)(n & 0xff);
            buf[1] = (byte)((n >> 8) & 0xff);
            buf[2] = (byte)((n >> 16) & 0xff);
            buf[3] = (byte)((n >> 24) & 0xff);
            buf[4] = (byte)((n >> 32) & 0xff);
            buf[5] = (byte)((n >> 40) & 0xff);
            buf[6] = (byte)((n >> 48) & 0xff);
            buf[7] = (byte)((n >> 56) & 0xff);
        }

        private long bytesToLong(byte[] bytes)
        {
            return
              ((bytes[7] & 0xffL) << 56) |
              ((bytes[6] & 0xffL) << 48) |
              ((bytes[5] & 0xffL) << 40) |
              ((bytes[4] & 0xffL) << 32) |
              ((bytes[3] & 0xffL) << 24) |
              ((bytes[2] & 0xffL) << 16) |
              ((bytes[1] & 0xffL) << 8) |
              ((bytes[0] & 0xffL));
        }

        byte[] varint64out = new byte[10];
        private void WriteVarint64(ulong n)
        {
            int idx = 0;
            while (true)
            {
                if ((n & ~(ulong)0x7FL) == 0)
                {
                    varint64out[idx++] = (byte)n;
                    break;
                }
                else
                {
                    varint64out[idx++] = ((byte)((n & 0x7F) | 0x80));
                    n >>= 7;
                }
            }
            databuffer.Write(varint64out, 0, idx);
        }

        byte[] i32buf = new byte[5];
        private void WriteVarint32(uint n)
        {
            int idx = 0;
            while (true)
            {
                if ((n & ~0x7F) == 0)
                {
                    i32buf[idx++] = (byte)n;
                    break;
                }
                else
                {
                    i32buf[idx++] = (byte)((n & 0x7F) | 0x80);
                    n >>= 7;
                }
            }
            databuffer.Write(i32buf, 0, idx);
        }

        private ulong ReadVarint64()
        {
            int shift = 0;
            ulong result = 0;
            while (true)
            {
                byte b = (byte)ReadByte();
                result |= (ulong)(b & 0x7f) << shift;
                if ((b & 0x80) != 0x80) break;
                shift += 7;
            }

            return result;
        }

        private uint ReadVarint32()
        {
            uint result = 0;
            int shift = 0;
            while (true)
            {
                byte b = (byte)ReadByte();
                result |= (uint)(b & 0x7f) << shift;
                if ((b & 0x80) != 0x80) break;
                shift += 7;
            }
            return result;
        }


        private ulong longToZigzag(long n)
        {
            return (ulong)(n << 1) ^ (ulong)(n >> 63);
        }

        private long zigzagToLong(ulong n)
        {
            return (long)(n >> 1) ^ (-(long)(n & 1));
        }

        private uint intToZigZag(int n)
        {
            return (uint)(n << 1) ^ (uint)(n >> 31);
        }

        private int zigzagToInt(uint n)
        {
            return (int)(n >> 1) ^ (-(int)(n & 1));
        }
    }

}
