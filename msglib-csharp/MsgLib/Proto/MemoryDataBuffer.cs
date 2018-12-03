using System;
using System.Collections.Generic;
using System.IO;
using System.Text;

namespace MsgLib
{
    internal class MemoryDataBuffer : IDataBuffer
    {
        internal MemoryStream Data { get; set; }

        internal MemoryDataBuffer()
        {
            this.Data = new MemoryStream();
        }
		internal MemoryDataBuffer(byte[] data)
			: this(data, false)
		{
		}
        internal MemoryDataBuffer(byte[] data, bool writable )
        {
            this.Data = new MemoryStream(data, writable);
        }

        internal MemoryDataBuffer(MemoryStream data)
        {
            this.Data = data;
        }

        int IDataBuffer.Read(byte[] buffer, int offset, int length)
        {
            return Data.Read(buffer, offset, length);
        }

        void IDataBuffer.Write(byte[] buffer, int offset, int length)
        {
            Data.Write(buffer, offset, length);
        }
    }
}
